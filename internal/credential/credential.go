// Package credential resolves API keys from the OS keychain.
// Resolution order: OS keychain > environment variable.
// The package knows nothing about providers — it stores and retrieves by name.
package credential

import (
	"fmt"
	"os"
	"strings"

	"github.com/zalando/go-keyring"
)

const serviceName = "ok"

// Resolve returns the secret for a given credential name.
// Resolution: keychain (ok:<name>) > env var (uppercased name + _API_KEY).
func Resolve(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("no credential name specified")
	}

	// 1. OS keychain
	if key, err := keyring.Get(serviceName, name); err == nil && key != "" {
		return key, nil
	}

	// 2. Environment variable: "anthropic-personal" → ANTHROPIC_PERSONAL_API_KEY
	envVar := toEnvVar(name)
	if key := os.Getenv(envVar); key != "" {
		return key, nil
	}

	return "", fmt.Errorf("no credential found for %q (checked keychain ok:%s + env %s)", name, name, envVar)
}

// Store saves a secret to the OS keychain under ok:<name>.
func Store(name, secret string) error {
	return keyring.Set(serviceName, name, secret)
}

// Delete removes a secret from the OS keychain.
func Delete(name string) error {
	return keyring.Delete(serviceName, name)
}

// Exists checks whether a credential is resolvable (keychain or env).
func Exists(name string) bool {
	_, err := Resolve(name)
	return err == nil
}

// toEnvVar converts a credential name to an env var name.
// "anthropic-personal" → "ANTHROPIC_PERSONAL_API_KEY"
func toEnvVar(name string) string {
	s := strings.ToUpper(name)
	s = strings.ReplaceAll(s, "-", "_")
	return s + "_API_KEY"
}
