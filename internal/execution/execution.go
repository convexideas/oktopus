// Package execution owns the session execution bounded context.
// Sessions, sandboxes, harnesses, assembly, and the launcher service.
package execution

import "context"

// Session represents a tracked agent execution.
type Session struct {
	ID        string `db:"id"`
	Agent     string `db:"agent"`
	Workspace string `db:"workspace"`
	Status    string `db:"status"`
	Title     string `db:"title"`
	CreatedAt string `db:"created_at"`
	EndedAt   string `db:"ended_at"`
}

// Workspace is a named collection of resource references.
type Workspace struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	CreatedAt string `db:"created_at"`
}

// WorkspaceRef is a typed reference bound to a workspace.
type WorkspaceRef struct {
	WorkspaceID string `db:"workspace_id"`
	Kind        string `db:"kind"`
	Slot        string `db:"slot"`
	Ref         string `db:"ref"`
	CreatedAt   string `db:"created_at"`
}

// Harness is the adapter interface for agent runtimes.
type Harness interface {
	Name() string
	BinaryName() string
	Start(ctx context.Context, cfg HarnessConfig) (HarnessSession, error)
}

// HarnessSession represents a running agent process.
type HarnessSession interface {
	ID() string
	Wait() error
	Stop() error
	ExitCode() int
	Output() string // captured stdout (non-interactive mode only)
}

// HarnessConfig holds the configuration for starting a harness.
type HarnessConfig struct {
	Workspace        string
	Env              map[string]string
	ProxyAddr        string
	Args             []string
	LogPath          string
	SystemPrompt     string
	SystemPromptMode string
	Model            string
	Tools            []string
	ExcludeTools     []string
	Task             string
	Files            map[string]string
	Internal         bool // if true, skip memory capture (prevents recursion)
}

// Store is the repository interface for sessions and workspaces.
type Store interface {
	CreateSession(ctx context.Context, s *Session) error
	CompleteSession(ctx context.Context, id, status string) error
	ListSessions(ctx context.Context, limit int) ([]Session, error)

	CreateWorkspace(ctx context.Context, ws *Workspace) error
	GetWorkspaceByName(ctx context.Context, name string) (*Workspace, error)
	ListWorkspaces(ctx context.Context) ([]Workspace, error)
	AddWorkspaceRef(ctx context.Context, ref *WorkspaceRef) error
	ListWorkspaceRefs(ctx context.Context, workspaceID string) ([]WorkspaceRef, error)
}
