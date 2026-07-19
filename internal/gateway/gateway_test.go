package gateway

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProxyCapturesAndForwards(t *testing.T) {
	// Mock upstream (pretend Anthropic)
	var gotAuth string
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("x-api-key")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"usage":{"input_tokens":100,"output_tokens":50},"content":[{"text":"hello"}]}`))
	}))
	defer mock.Close()

	// Gateway pointing at mock as upstream
	meter := NewMeter("test-session")
	provider := ProviderConfig{
		Name:    "anthropic",
		BaseURL: mock.URL,
		APIKey:  "sk-real-secret-key",
	}
	gw := New(meter, provider)
	if err := gw.Start(); err != nil {
		t.Fatal(err)
	}
	defer gw.Stop()

	// Agent calls our gateway (as if ANTHROPIC_BASE_URL=http://localhost:port)
	body := strings.NewReader(`{"model":"claude-sonnet-4-20250514","messages":[{"role":"user","content":"hi"}]}`)
	req, _ := http.NewRequest("POST", gw.BaseURL()+"/v1/messages", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", "ok-gateway") // sentinel — should be stripped

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	// Verify credential injection
	if gotAuth != "sk-real-secret-key" {
		t.Errorf("expected real key injected, got %q", gotAuth)
	}

	// Verify capture
	if meter.Count() != 1 {
		t.Fatalf("expected 1 capture, got %d", meter.Count())
	}

	tokensIn, tokensOut, _ := meter.Totals()
	if tokensIn != 100 || tokensOut != 50 {
		t.Errorf("expected 100/50 tokens, got %d/%d", tokensIn, tokensOut)
	}

	caps := meter.Flush()
	if caps[0].Model != "claude-sonnet-4-20250514" {
		t.Errorf("expected model claude-sonnet-4-20250514, got %q", caps[0].Model)
	}
	if caps[0].Provider != "anthropic" {
		t.Errorf("expected provider anthropic, got %q", caps[0].Provider)
	}
}

func TestProxyStripsAgentAuth(t *testing.T) {
	// Verify agent-supplied auth is stripped, not forwarded
	var gotAuthHeader, gotAPIKeyHeader string
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthHeader = r.Header.Get("Authorization")
		gotAPIKeyHeader = r.Header.Get("x-api-key")
		w.Write([]byte(`{}`))
	}))
	defer mock.Close()

	meter := NewMeter("test")
	gw := New(meter, ProviderConfig{Name: "openai", BaseURL: mock.URL, APIKey: "sk-real"})
	gw.Start()
	defer gw.Stop()

	req, _ := http.NewRequest("POST", gw.BaseURL()+"/v1/chat/completions", strings.NewReader(`{"model":"gpt-4o"}`))
	req.Header.Set("Authorization", "Bearer agent-sentinel-should-be-stripped")
	req.Header.Set("x-api-key", "also-should-be-stripped")

	http.DefaultClient.Do(req)

	// For openai provider, should get Bearer with real key
	if gotAuthHeader != "Bearer sk-real" {
		t.Errorf("expected 'Bearer sk-real', got %q", gotAuthHeader)
	}
	// x-api-key should be stripped (not forwarded for openai)
	if gotAPIKeyHeader != "" {
		t.Errorf("expected empty x-api-key, got %q", gotAPIKeyHeader)
	}
}

func TestEstimateCost(t *testing.T) {
	cost := estimateCost("anthropic", "claude-sonnet-4-20250514", 1000, 500)
	if cost <= 0 {
		t.Errorf("expected positive cost for sonnet, got %f", cost)
	}

	cost = estimateCost("openai", "gpt-4o", 1000, 500)
	if cost <= 0 {
		t.Errorf("expected positive cost for gpt-4o, got %f", cost)
	}

	cost = estimateCost("ollama", "llama3", 1000, 500)
	if cost != 0 {
		t.Errorf("expected zero cost for unknown model, got %f", cost)
	}
}
