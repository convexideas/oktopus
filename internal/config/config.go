// Package config provides application configuration loaded via koanf.
// Layering: defaults < config file (~/.ok/config.yaml) < environment variables (OK_ prefix).
package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

const (
	appDirName = "ok"
	ConfigFile = "config.yaml"
	DBFile     = "oktopus.db"
)

// APIProtocol is a wire format the gateway knows how to proxy.
// This is the ONLY thing the codebase knows about providers.
type APIProtocol string

const (
	ProtocolAnthropic APIProtocol = "anthropic"
	ProtocolOpenAI    APIProtocol = "openai"
)

// APIProvider is a named inference endpoint from user config.
type APIProvider struct {
	Protocol   APIProtocol `koanf:"protocol"`
	BaseURL    string      `koanf:"base_url"`
	Credential string      `koanf:"credential"` // keychain key (ok:<credential>)
}

// SandboxProvider is a named sandbox backend from user config.
type SandboxProvider struct {
	Type       string         `koanf:"type"`       // auto, seatbelt, bwrap, container, vm, process
	Credential string         `koanf:"credential"` // keychain key for remote providers (modal, fly, daytona)
	Config     map[string]any `koanf:"config"`     // provider-specific
}

// Defaults holds user-facing defaults for resolution.
type Defaults struct {
	APIProvider     string `koanf:"api_provider"`     // key into APIProviders
	SandboxProvider string `koanf:"sandbox_provider"` // key into SandboxProviders
	Harness         string `koanf:"harness"`
	Model           string `koanf:"model"`
}

// Config holds all resolved application configuration.
type Config struct {
	DBPath  string `koanf:"db_path"`
	LogDir  string `koanf:"log_dir"`
	HomeDir string `koanf:"home_dir"`

	// Provider registries — loaded from config, not hardcoded
	APIProviders     map[string]APIProvider     `koanf:"api_providers"`
	SandboxProviders map[string]SandboxProvider `koanf:"sandbox_providers"`
	Defaults         Defaults                   `koanf:"defaults"`

	// Legacy — kept for backward compat during migration
	ProxyAddr string `koanf:"proxy_addr"`
}

// ResolveAPIProvider returns the named API provider config.
// Falls back to the default if name is empty.
func (c *Config) ResolveAPIProvider(name string) (string, *APIProvider) {
	if name == "" {
		name = c.Defaults.APIProvider
	}
	if name == "" {
		return "", nil
	}
	if p, ok := c.APIProviders[name]; ok {
		return name, &p
	}
	return "", nil
}

// ResolveSandboxProvider returns the named sandbox provider config.
// Falls back to the default if name is empty.
func (c *Config) ResolveSandboxProvider(name string) (string, *SandboxProvider) {
	if name == "" {
		name = c.Defaults.SandboxProvider
	}
	if name == "" {
		// Implicit default: local with auto-detect
		return "local", &SandboxProvider{Type: "auto"}
	}
	if p, ok := c.SandboxProviders[name]; ok {
		return name, &p
	}
	// Fallback: local/auto
	return "local", &SandboxProvider{Type: "auto"}
}

// Home returns the base directory for all ok state.
// Resolution: $OK_HOME > $XDG_CONFIG_HOME/ok > ~/.ok
func Home() string {
	if dir := os.Getenv("OK_HOME"); dir != "" {
		return dir
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, appDirName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "."+appDirName)
}

// Load resolves configuration by layering defaults, config file, and env vars.
func Load() (*Config, error) {
	k := koanf.New(".")

	okHome := Home()

	defaults := Config{
		DBPath:  filepath.Join(okHome, DBFile),
		HomeDir: okHome,
	}

	// 1. Defaults
	if err := k.Load(structs.Provider(defaults, "koanf"), nil); err != nil {
		return nil, err
	}

	// 2. Config file (optional — not an error if missing)
	cfgPath := filepath.Join(okHome, ConfigFile)
	if _, err := os.Stat(cfgPath); err == nil {
		_ = k.Load(file.Provider(cfgPath), yaml.Parser())
	}

	// 3. Environment variables: OK_DB_PATH → db_path
	_ = k.Load(env.Provider("OK_", ".", func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "OK_"))
	}), nil)

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
