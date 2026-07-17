// Package runtime owns the live execution context — sessions, sandboxes, harnesses, and assembly.
package runtime

import "context"

// Session is an audit record of a block of work. Immutable after completion.
// Continuity is provided by the sandbox (harness resumes natively).
type Session struct {
	ID        string `db:"id"`
	Agent     string `db:"agent"`
	Workspace string `db:"workspace"`
	Status    string `db:"status"`
	Title     string `db:"title"`
	CreatedAt string `db:"created_at"`
	EndedAt   string `db:"ended_at"`
}

// Workspace is a named scope of work. Declares I/O, connectors, constraints.
// ponytail: lives in runtime/ for now. Moves to its own package when store is split.
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

// SessionStore persists session records.
type SessionStore interface {
	CreateSession(ctx context.Context, s *Session) error
	CompleteSession(ctx context.Context, id, status string) error
	ListSessions(ctx context.Context, limit int) ([]Session, error)

	CreateWorkspace(ctx context.Context, ws *Workspace) error
	GetWorkspaceByName(ctx context.Context, name string) (*Workspace, error)
	ListWorkspaces(ctx context.Context) ([]Workspace, error)
	AddWorkspaceRef(ctx context.Context, ref *WorkspaceRef) error
	ListWorkspaceRefs(ctx context.Context, workspaceID string) ([]WorkspaceRef, error)
}
