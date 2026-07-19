package gateway

import (
	"encoding/json"
	"strings"
)

// detectProvider identifies LLM API providers by host.
func detectProvider(host string) string {
	switch {
	case strings.Contains(host, "anthropic.com"):
		return "anthropic"
	case strings.Contains(host, "openai.com"):
		return "openai"
	case strings.HasPrefix(host, "localhost"), strings.HasPrefix(host, "127.0.0.1"):
		return "ollama"
	default:
		return ""
	}
}

// extractModel pulls the model name from a request body.
func extractModel(body []byte) string {
	var req struct {
		Model string `json:"model"`
	}
	json.Unmarshal(body, &req)
	return req.Model
}

// extractUsage pulls token counts from a response body.
// Handles both Anthropic and OpenAI response formats.
func extractUsage(provider string, body []byte) (tokensIn, tokensOut int) {
	var resp struct {
		Usage struct {
			InputTokens      int `json:"input_tokens"`      // Anthropic
			OutputTokens     int `json:"output_tokens"`     // Anthropic
			PromptTokens     int `json:"prompt_tokens"`     // OpenAI
			CompletionTokens int `json:"completion_tokens"` // OpenAI
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, 0
	}
	switch provider {
	case "anthropic":
		return resp.Usage.InputTokens, resp.Usage.OutputTokens
	case "openai":
		return resp.Usage.PromptTokens, resp.Usage.CompletionTokens
	default:
		return 0, 0
	}
}

// estimateCost returns a rough USD cost estimate.
// ponytail: hardcoded pricing. Replace with configurable pricing table when policy needs it.
func estimateCost(provider, model string, tokensIn, tokensOut int) float64 {
	var inRate, outRate float64 // per million tokens
	switch {
	case strings.Contains(model, "opus"):
		inRate, outRate = 15.0, 75.0
	case strings.Contains(model, "sonnet"):
		inRate, outRate = 3.0, 15.0
	case strings.Contains(model, "haiku"):
		inRate, outRate = 0.25, 1.25
	case strings.Contains(model, "gpt-4o"):
		inRate, outRate = 2.5, 10.0
	case strings.Contains(model, "gpt-4"):
		inRate, outRate = 30.0, 60.0
	default:
		return 0
	}
	return (float64(tokensIn)*inRate + float64(tokensOut)*outRate) / 1_000_000
}
