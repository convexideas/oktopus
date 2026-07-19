package gateway

import (
	"encoding/json"
	"strings"
)

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
		// Try both formats
		if resp.Usage.InputTokens > 0 {
			return resp.Usage.InputTokens, resp.Usage.OutputTokens
		}
		return resp.Usage.PromptTokens, resp.Usage.CompletionTokens
	}
}

// estimateCost returns a rough USD cost estimate.
// ponytail: hardcoded pricing. Replace with configurable pricing table.
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
