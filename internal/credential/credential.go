// Package credential resolves API keys for LLM providers.
// Resolution order: OS keychain > environment variable > file.
package credential

import (
	"fmt"
	"os"

	"github.com/zalando/go-keyring"
)

const serviceName = "oktopus"

// Provider names used as keychain keys.
const (
	Anthropic = "anthropic"
	OpenAI    = "openai"
	Ollama    = "ollama"
)

// envVars maps provider names to their conventional env var.
var envVars = map[string]string{
	Anthropic: "ANTHROPIC_API_KEY",
	OpenAI:    "OPENAI_API_KEY",
}

// Resolve returns the API key for a provider.
// Resolution: keychain > env var.
func Resolve(provider string) (string, error) {
	// 1. OS keychain
	if key, err := keyring.Get(serviceName, provider); err == nil && key != "" {
		return key, nil
	}

	// 2. Environment variable
	if envVar, ok := envVars[provider]; ok {
		if key := os.Getenv(envVar); key != "" {
			return key, nil
		}
	}

	return "", fmt.Errorf("no credential found for provider %q (checked keychain + env)", provider)
}

// Store saves an API key to the OS keychain.
func Store(provider, apiKey string) error {
	return keyring.Set(serviceName, provider, apiKey)
}

// Delete removes an API key from the OS keychain.
func Delete(provider string) error {
	return keyring.Delete(serviceName, provider)
}

// List returns which providers have stored credentials.
func List() []string {
	var found []string
	for _, p := range []string{Anthropic, OpenAI, Ollama} {
		if _, err := keyring.Get(serviceName, p); err == nil {
			found = append(found, p)
		}
	}
	return found
}
