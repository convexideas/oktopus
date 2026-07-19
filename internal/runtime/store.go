package runtime

import "context"

// SessionStore persists session lifecycle records.
type SessionStore interface {
	CreateSession(ctx context.Context, sess *Session) error
	CompleteSession(ctx context.Context, id, status string) error
	ListSessions(ctx context.Context, limit int) ([]Session, error)
}

// WorkspaceStore persists workspace definitions and references.
type WorkspaceStore interface {
	CreateWorkspace(ctx context.Context, ws *Workspace) error
	GetWorkspaceByName(ctx context.Context, name string) (*Workspace, error)
	ListWorkspaces(ctx context.Context) ([]Workspace, error)
	AddWorkspaceRef(ctx context.Context, ref *WorkspaceRef) error
	ListWorkspaceRefs(ctx context.Context, workspaceID string) ([]WorkspaceRef, error)
}
