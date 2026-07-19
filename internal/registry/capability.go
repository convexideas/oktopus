// Package registry owns the capability registry bounded context.
// Capabilities are versioned, importable items: personas, skills, tools, workflows.
package registry

import "context"

// Capability is an item in the registry.
type Capability struct {
	ID           string `db:"id"`
	Kind         string `db:"kind"`
	Name         string `db:"name"`
	Version      string `db:"version"`
	Description  string `db:"description"`
	Author       string `db:"author"`
	SourceType   string `db:"source_type"`
	Scope        string `db:"scope"`
	Status       string `db:"status"`
	ManifestJSON string `db:"manifest_json"`
	Hash         string `db:"hash"`
	ManifestPath string `db:"manifest_path"`
	CreatedAt    string `db:"created_at"`
	UpdatedAt    string `db:"updated_at"`
}

// PersonaSpec is the runtime configuration extracted from a persona manifest.
type PersonaSpec struct {
	SystemPrompt string
	PromptMode   string // "append" or "replace"
	Skills       []string
	OutputFormat string
	Native       map[string]NativeConfig
}

// NativeConfig holds harness-specific overrides defined in a persona.
type NativeConfig struct {
	Files    map[string]string
	Flags    []string
	Settings map[string]any
}

// Store is the repository interface for capabilities.
type Store interface {
	Register(ctx context.Context, c *Capability) error
	FindCapability(ctx context.Context, kind, name, version string) (*Capability, error)
	FindCapabilityByName(ctx context.Context, kind, name string) (*Capability, error)
	ListCapabilitiesByKind(ctx context.Context, kind string) ([]Capability, error)
	ListCapabilities(ctx context.Context) ([]Capability, error)
}
