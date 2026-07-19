// Package gateway provides an HTTP forward proxy that sits between
// agent sandboxes and LLM API providers.
//
// Phase 1: plain HTTP capture. CONNECT tunnels pass through opaquely.
// Phase 2: TLS interception (CA cert).
// Phase 3: credential injection, inference routing, policy enforcement.
package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Proxy is an HTTP forward proxy that captures LLM API traffic.
type Proxy struct {
	Meter *Meter
	Addr  string // populated after Start: "127.0.0.1:<port>"

	server   *http.Server
	listener net.Listener
}

// New creates a proxy wired to the given meter.
func New(meter *Meter) *Proxy {
	return &Proxy{Meter: meter}
}

// Start begins listening on a random local port.
func (p *Proxy) Start() error {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("gateway listen: %w", err)
	}
	p.listener = ln
	p.Addr = ln.Addr().String()
	p.server = &http.Server{Handler: p}
	go p.server.Serve(ln)
	return nil
}

// Stop shuts down the proxy gracefully.
func (p *Proxy) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return p.server.Shutdown(ctx)
}

// ServeHTTP routes requests: CONNECT → tunnel, plain HTTP → capture.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.handleConnect(w, r)
		return
	}
	p.handleHTTP(w, r)
}

// handleConnect tunnels HTTPS without interception (phase 1 — passthrough).
func (p *Proxy) handleConnect(w http.ResponseWriter, r *http.Request) {
	dest, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijack not supported", http.StatusInternalServerError)
		dest.Close()
		return
	}

	w.WriteHeader(http.StatusOK)
	client, _, err := hijacker.Hijack()
	if err != nil {
		dest.Close()
		return
	}

	go transfer(dest, client)
	go transfer(client, dest)
}

func transfer(dst, src net.Conn) {
	io.Copy(dst, src)
	dst.Close()
}

// handleHTTP proxies plain HTTP with full body capture.
func (p *Proxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Read request body
	var reqBody []byte
	if r.Body != nil {
		reqBody, _ = io.ReadAll(r.Body)
		r.Body.Close()
	}

	// Forward request
	outReq, _ := http.NewRequestWithContext(r.Context(), r.Method, r.URL.String(), bytes.NewReader(reqBody))
	outReq.Header = r.Header.Clone()

	resp, err := http.DefaultTransport.RoundTrip(outReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Read response body
	respBody, _ := io.ReadAll(resp.Body)

	// Write back to client
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)

	// Capture if LLM API call
	provider := detectProvider(r.Host)
	if provider == "" {
		return
	}

	tokensIn, tokensOut := extractUsage(provider, respBody)
	model := extractModel(reqBody)

	cap := Capture{
		ID:        uuid.New().String(),
		SessionID: p.Meter.sessionID,
		Timestamp: time.Now().UTC(),
		Provider:  provider,
		Model:     model,
		Endpoint:  r.URL.Path,
		TokensIn:  tokensIn,
		TokensOut: tokensOut,
		Cost:      estimateCost(provider, model, tokensIn, tokensOut),
		Duration:  time.Since(start),
		Request:   reqBody,
		Response:  respBody,
	}
	p.Meter.Record(cap)
}
