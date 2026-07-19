// Package gateway provides a local reverse proxy that agents talk to
// instead of real LLM API endpoints. The agent is configured with
// ANTHROPIC_BASE_URL=http://localhost:<port> (or similar) so all
// inference traffic flows through us.
//
// The gateway:
//   - Captures full request/response (structured messages, tool calls)
//   - Meters token usage and cost
//   - Injects real credentials (agent never sees them)
//   - Routes to the configured upstream provider
//
// No MITM, no CA certs, no trust store manipulation.
// Inspired by OpenShell's inference.local pattern.
package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/convexideas/oktopus/internal/config"
	"github.com/google/uuid"
)

// ProviderConfig describes how to reach an upstream LLM provider.
// All fields come from user config — nothing is hardcoded in the gateway.
type ProviderConfig struct {
	Name     string             // user-defined name (for capture logging)
	Protocol config.APIProtocol // wire format: determines auth header injection
	BaseURL  string             // upstream endpoint
	APIKey   string             // real credential — injected by gateway, never exposed to agent
}

// Proxy is a local reverse proxy that captures and meters LLM API traffic.
type Proxy struct {
	Meter    *Meter
	Provider ProviderConfig
	Addr     string // populated after Start: "127.0.0.1:<port>"

	server   *http.Server
	listener net.Listener
}

// New creates a proxy that routes to the given provider.
func New(meter *Meter, provider ProviderConfig) *Proxy {
	return &Proxy{Meter: meter, Provider: provider}
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

// BaseURL returns the URL agents should use as their API base.
func (p *Proxy) BaseURL() string {
	return "http://" + p.Addr
}

// ServeHTTP handles all requests from the agent.
// Strips agent-supplied auth, injects real credentials, forwards to upstream.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Read request body
	var reqBody []byte
	if r.Body != nil {
		reqBody, _ = io.ReadAll(r.Body)
		r.Body.Close()
	}

	// Build upstream URL
	upstreamURL := strings.TrimRight(p.Provider.BaseURL, "/") + "/" + strings.TrimLeft(r.URL.Path, "/")
	if r.URL.RawQuery != "" {
		upstreamURL += "?" + r.URL.RawQuery
	}

	// Forward request
	outReq, _ := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL, bytes.NewReader(reqBody))

	// Copy headers, strip agent auth
	for k, vv := range r.Header {
		lower := strings.ToLower(k)
		if lower == "authorization" || lower == "x-api-key" {
			continue
		}
		for _, v := range vv {
			outReq.Header.Add(k, v)
		}
	}

	// Inject real credentials
	p.injectAuth(outReq)

	resp, err := http.DefaultTransport.RoundTrip(outReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Read response
	respBody, _ := io.ReadAll(resp.Body)

	// Write back to agent
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)

	// Capture + meter
	tokensIn, tokensOut := extractUsage(string(p.Provider.Protocol), respBody)
	model := extractModel(reqBody)

	cap := Capture{
		ID:        uuid.New().String(),
		SessionID: p.Meter.sessionID,
		Timestamp: time.Now().UTC(),
		Provider:  p.Provider.Name,
		Model:     model,
		Endpoint:  r.URL.Path,
		TokensIn:  tokensIn,
		TokensOut: tokensOut,
		Cost:      estimateCost(string(p.Provider.Protocol), model, tokensIn, tokensOut),
		Duration:  time.Since(start),
		Request:   reqBody,
		Response:  respBody,
	}
	p.Meter.Record(cap)
}

// injectAuth adds the appropriate auth header for the upstream provider.
func (p *Proxy) injectAuth(r *http.Request) {
	if p.Provider.APIKey == "" {
		return
	}
	switch p.Provider.Protocol {
	case config.ProtocolAnthropic:
		r.Header.Set("x-api-key", p.Provider.APIKey)
	default:
		// OpenAI and compatible (OpenRouter, Ollama, etc.)
		r.Header.Set("Authorization", "Bearer "+p.Provider.APIKey)
	}
}
