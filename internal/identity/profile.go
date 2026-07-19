// Package identity owns the user identity bounded context.
// Profile is the aggregate root — a user's complete runtime identity.
package identity

import "context"

// Store persists and retrieves user profiles.
type Store interface {
	GetProfile(ctx context.Context, userID string) (*Profile, error)
	UpsertProfile(ctx context.Context, profile *Profile) error
}

// Profile is the user's runtime identity.
type Profile struct {
	ID        string `db:"id"`
	UserID    string `db:"user_id"`
	DataJSON  string `db:"config_json"` // serialized profileData
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`

	// Hydrated fields (not stored directly — derived from DataJSON)
	Preferences   Preferences   `db:"-"`
	HarnessConfig HarnessConfig `db:"-"`
}

// profileData is the internal serialization envelope.
type profileData struct {
	Preferences   Preferences   `json:"preferences"`
	HarnessConfig HarnessConfig `json:"harness_config"`
}

// Preferences holds portable user choices — what capabilities to use.
// These are harness-agnostic; adapters translate them to harness-specific config.
type Preferences struct {
	DefaultModel   string      `json:"default_model,omitempty"`
	DefaultRuntime string      `json:"default_runtime,omitempty"`
	Extensions     []Extension `json:"extensions,omitempty"`
	Skills         []string    `json:"skills,omitempty"`
	ToolsExclude   []string    `json:"tools_exclude,omitempty"` // generic names, mapped by adapter
}

// Extension represents an MCP server or tool extension.
// Portable across harnesses — assembler wires them per-harness.
type Extension struct {
	Name    string            `json:"name"`
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// HarnessConfig holds per-harness specific settings that have no generic equivalent.
type HarnessConfig struct {
	Pi    *PiConfig    `json:"pi,omitempty"`
	Codex *CodexConfig `json:"codex,omitempty"`
	Kiro  *KiroConfig  `json:"kiro,omitempty"`
}

// PiConfig holds Pi-specific settings.
type PiConfig struct {
	ApprovalMode string      `json:"approval_mode,omitempty"`
	Extensions   []Extension `json:"extensions,omitempty"` // Pi-only extensions
}

// CodexConfig holds Codex-specific settings.
type CodexConfig struct {
	ApprovalMode string      `json:"approval_mode,omitempty"`
	Extensions   []Extension `json:"extensions,omitempty"` // Codex-only extensions
}

// KiroConfig holds Kiro-specific settings.
type KiroConfig struct {
	SteeringFiles []string    `json:"steering_files,omitempty"`
	Extensions    []Extension `json:"extensions,omitempty"` // Kiro-only extensions
}


