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

// Config holds all resolved application configuration.
type Config struct {
	DBPath    string `koanf:"db_path"`
	LogDir    string `koanf:"log_dir"`
	ProxyAddr string `koanf:"proxy_addr"`
	HomeDir   string `koanf:"home_dir"`
}

// resolveHome returns the base directory for all ok state.
// Resolution: $OK_HOME > $XDG_CONFIG_HOME/ok > ~/.ok
func resolveHome() string {
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

	okHome := resolveHome()

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
