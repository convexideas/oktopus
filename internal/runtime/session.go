// Package runtime owns the live execution context — sessions, sandboxes, harnesses, and assembly.
package runtime

import "time"

// SessionState represents the lifecycle state of a session.
type SessionState string

const (
	StateCreated   SessionState = "created"
	StateRunning   SessionState = "running"
	StateCompleted SessionState = "completed"
	StateFailed    SessionState = "failed"
	StateCancelled SessionState = "cancelled"
)

// Session is an audit record of a block of work. Immutable after completion.
// Continuity is provided by the sandbox (harness resumes natively).
type Session struct {
	ID        string `db:"id"`
	Agent     string `db:"agent"`
	Workspace string `db:"workspace"`
	Sandbox   string `db:"sandbox"`
	Status    string `db:"status"`
	Title     string `db:"title"`
	CreatedAt string `db:"created_at"`
	EndedAt   string `db:"ended_at"`
}

// NewSession creates a session in the "created" state.
func NewSession(id, agent, workspace, sandbox string) *Session {
	now := time.Now().UTC().Format(time.RFC3339)
	return &Session{
		ID:        id,
		Agent:     agent,
		Workspace: workspace,
		Sandbox:   sandbox,
		Title:     agent,
		Status:    string(StateCreated),
		CreatedAt: now,
	}
}

// Start transitions the session to running.
func (s *Session) Start() {
	s.Status = string(StateRunning)
}

// Complete transitions the session to completed.
func (s *Session) Complete() {
	s.Status = string(StateCompleted)
	s.EndedAt = time.Now().UTC().Format(time.RFC3339)
}

// Fail transitions the session to failed.
func (s *Session) Fail() {
	s.Status = string(StateFailed)
	s.EndedAt = time.Now().UTC().Format(time.RFC3339)
}

// Cancel transitions the session to cancelled.
func (s *Session) Cancel() {
	s.Status = string(StateCancelled)
	s.EndedAt = time.Now().UTC().Format(time.RFC3339)
}

// Workspace is a named scope of work. Declares I/O, connectors, constraints.
// ponytail: lives in runtime/ for now. Moves to its own package when store is split.
type Workspace struct {
	ID         string `db:"id"`
	Name       string `db:"name"`
	ConfigJSON string `db:"config_json"` // serialized WorkspaceConfig
	CreatedAt  string `db:"created_at"`
}

// WorkspaceConfig holds workspace-level configuration including sandbox defaults.
type WorkspaceConfig struct {
	Sandbox SandboxConfig `json:"sandbox,omitempty"`
}

// SandboxConfig declares how sandboxes are created for this workspace.
type SandboxConfig struct {
	Provider string         `json:"provider,omitempty"` // "local", "modal", "fly", "daytona"
	Type     string         `json:"type,omitempty"`     // "process", "container", "vm"
	Config   map[string]any `json:"config,omitempty"`   // type-specific (image, resources, flavor)
}

// WorkspaceRef is a typed reference bound to a workspace.
type WorkspaceRef struct {
	WorkspaceID string `db:"workspace_id"`
	Kind        string `db:"kind"`
	Slot        string `db:"slot"`
	Ref         string `db:"ref"`
	CreatedAt   string `db:"created_at"`
}


