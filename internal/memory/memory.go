// Package memory owns the memory bounded context.
// Episodes are session captures: raw output and summaries.
// Summaries provide continuity between sessions — injected as context into future runs.
package memory

import "context"

// Episode is a capture from a session.
type Episode struct {
	ID          string `db:"id"`
	SessionID   string `db:"session_id"`
	WorkspaceID string `db:"workspace_id"`
	Source      string `db:"source"` // "stdout" or "summary"
	Content     string `db:"content"`
	CapturedAt  string `db:"captured_at"`
}

// Store persists and retrieves episodes.
type Store interface {
	Save(ctx context.Context, ep *Episode) error
	ListByWorkspace(ctx context.Context, workspaceID string, limit int) ([]Episode, error)
	ListSummariesByWorkspace(ctx context.Context, workspaceID string, limit int) ([]Episode, error)
	ListRecent(ctx context.Context, limit int) ([]Episode, error)
}
