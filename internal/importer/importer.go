// Package importer handles discovering installed harnesses and importing
// their settings + session history into Oktopus.
//
// Architecture: Base provides generic machinery (discover path, read config,
// walk sessions). Each harness embeds Base and overrides only what's specific
// (e.g., how to parse its session file format).
package importer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/convexideas/oktopus/internal/identity"
)

// Harness is the interface all importable harnesses satisfy.
type Harness interface {
	Name() string
	Discover() (*Discovery, error)
	Settings() (*identity.Preferences, *identity.HarnessConfig, error)
	Sessions(limit int) ([]Episode, error)
}

// Discovery is the result of probing for a harness.
type Discovery struct {
	Found    bool
	Path     string
	Model    string
	Provider string
	Packages []string
	Sessions int
}

// Episode is an imported session ready for storage.
type Episode struct {
	Timestamp string
	CWD       string
	Content   string
}

// FieldMap tells the generic config reader which keys to extract.
type FieldMap struct {
	Model    string // key in config that holds default model
	Provider string // key that holds default provider
	Packages string // key that holds list of packages/extensions
}

// SessionParser extracts readable content from a session file.
// Each harness provides its own implementation.
type SessionParser func(path string) (string, error)

// Base provides generic import machinery. Embed in harness implementations.
type Base struct {
	name          string
	homeDir       string // relative to ~
	configFile    string
	fieldMap      FieldMap
	sessionDir    string
	sessionExt    string
	sessionParser SessionParser
	root          string // resolved absolute path (set during Discover)
}

func (b *Base) Name() string { return b.name }

func (b *Base) Discover() (*Discovery, error) {
	home, _ := os.UserHomeDir()
	b.root = filepath.Join(home, b.homeDir)

	d := &Discovery{}
	if _, err := os.Stat(b.root); err != nil {
		return d, nil
	}
	d.Found = true
	d.Path = b.root

	// Generic config reading
	if b.configFile != "" {
		if cfg := b.readConfig(); cfg != nil {
			d.Model = stringField(cfg, b.fieldMap.Model)
			d.Provider = stringField(cfg, b.fieldMap.Provider)
			d.Packages = sliceField(cfg, b.fieldMap.Packages)
		}
	}

	// Generic session counting
	if b.sessionDir != "" {
		d.Sessions = b.countSessions()
	}

	return d, nil
}

func (b *Base) Settings() (*identity.Preferences, *identity.HarnessConfig, error) {
	cfg := b.readConfig()
	if cfg == nil {
		return nil, nil, nil
	}

	prefs := &identity.Preferences{
		DefaultRuntime: b.name,
		DefaultModel:   stringField(cfg, b.fieldMap.Model),
	}

	for _, pkg := range sliceField(cfg, b.fieldMap.Packages) {
		if ext := packageToExtension(pkg); ext != nil {
			prefs.Extensions = append(prefs.Extensions, *ext)
		}
	}

	return prefs, nil, nil
}

func (b *Base) Sessions(limit int) ([]Episode, error) {
	if b.sessionParser == nil {
		return nil, nil
	}

	sessPath := filepath.Join(b.root, b.sessionDir)
	entries, err := os.ReadDir(sessPath)
	if err != nil {
		return nil, err
	}

	var episodes []Episode
	count := 0

	for _, entry := range entries {
		if !entry.IsDir() || count >= limit {
			continue
		}

		// Skip our own sandbox sessions and system temp dirs
		name := entry.Name()
		if strings.Contains(name, "ok-sandbox") || strings.Contains(name, "private-var-folders") {
			continue
		}

		wsDir := filepath.Join(sessPath, name)
		cwd := decodeDirName(name)

		files, _ := os.ReadDir(wsDir)
		for _, f := range files {
			if !strings.HasSuffix(f.Name(), b.sessionExt) || count >= limit {
				continue
			}

			content, err := b.sessionParser(filepath.Join(wsDir, f.Name()))
			if err != nil || content == "" {
				continue
			}

			episodes = append(episodes, Episode{
				Timestamp: extractTimestamp(f.Name()),
				CWD:       cwd,
				Content:   content,
			})
			count++
		}
	}

	return episodes, nil
}

// --- internal helpers ---

func (b *Base) readConfig() map[string]any {
	data, err := os.ReadFile(filepath.Join(b.root, b.configFile))
	if err != nil {
		return nil
	}
	var cfg map[string]any
	json.Unmarshal(data, &cfg)
	return cfg
}

func (b *Base) countSessions() int {
	sessPath := filepath.Join(b.root, b.sessionDir)
	count := 0
	entries, _ := os.ReadDir(sessPath)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		files, _ := os.ReadDir(filepath.Join(sessPath, entry.Name()))
		for _, f := range files {
			if strings.HasSuffix(f.Name(), b.sessionExt) {
				count++
			}
		}
	}
	return count
}

func stringField(cfg map[string]any, key string) string {
	if key == "" {
		return ""
	}
	if v, ok := cfg[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func sliceField(cfg map[string]any, key string) []string {
	if key == "" {
		return nil
	}
	if v, ok := cfg[key]; ok {
		if arr, ok := v.([]any); ok {
			var result []string
			for _, item := range arr {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return nil
}

func decodeDirName(name string) string {
	p := strings.ReplaceAll(name, "--", "/")
	p = strings.TrimSuffix(p, "/")
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

func extractTimestamp(filename string) string {
	if idx := strings.Index(filename, "_"); idx > 0 {
		return filename[:idx]
	}
	return ""
}

func packageToExtension(pkg string) *identity.Extension {
	if strings.HasPrefix(pkg, "npm:") {
		name := strings.TrimPrefix(pkg, "npm:")
		displayName := name
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			displayName = name[idx+1:]
		}
		return &identity.Extension{
			Name:    displayName,
			Command: "npx",
			Args:    []string{"-y", name},
		}
	}
	if strings.HasPrefix(pkg, "git:") {
		repo := strings.TrimPrefix(pkg, "git:")
		parts := strings.Split(repo, "/")
		return &identity.Extension{
			Name:    parts[len(parts)-1],
			Command: "git:" + repo,
		}
	}
	return nil
}
