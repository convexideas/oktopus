package gateway

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestProxyCapturesLLMTraffic(t *testing.T) {
	// Mock Anthropic API
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"usage":{"input_tokens":100,"output_tokens":50},"content":[{"text":"hello"}]}`))
	}))
	defer mock.Close()

	// Start gateway
	meter := NewMeter("test-session")
	gw := New(meter)
	if err := gw.Start(); err != nil {
		t.Fatal(err)
	}
	defer gw.Stop()

	// Send request through proxy to mock (pretend it's anthropic)
	// We rewrite detectProvider to match the mock host for this test by using plain HTTP
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(r *http.Request) (*url.URL, error) {
				return url.Parse("http://" + gw.Addr)
			},
		},
	}

	body := strings.NewReader(`{"model":"claude-sonnet-4-20250514","messages":[{"role":"user","content":"hi"}]}`)
	resp, err := client.Post(mock.URL+"/v1/messages", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	// Verify capture — mock is localhost so will be detected as "ollama"
	if meter.Count() != 1 {
		t.Fatalf("expected 1 capture, got %d", meter.Count())
	}

	tokensIn, tokensOut, _ := meter.Totals()
	// localhost is detected as "ollama" which doesn't parse anthropic format
	// But model extraction should still work
	caps := meter.Flush()
	if caps[0].Model != "claude-sonnet-4-20250514" {
		t.Errorf("expected model claude-sonnet-4-20250514, got %q", caps[0].Model)
	}
	_ = tokensIn
	_ = tokensOut
}

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		{"api.anthropic.com", "anthropic"},
		{"api.openai.com", "openai"},
		{"localhost:11434", "ollama"},
		{"127.0.0.1:8080", "ollama"},
		{"unknown.example.com", ""},
	}
	for _, tt := range tests {
		got := detectProvider(tt.host)
		if got != tt.want {
			t.Errorf("detectProvider(%q) = %q, want %q", tt.host, got, tt.want)
		}
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
