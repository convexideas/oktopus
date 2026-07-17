package sqlite

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/convexideas/oktopus/internal/runtime"
	"github.com/google/uuid"
)

var wsCols = []string{"id", "name", "created_at"}
var wsRefCols = []string{"workspace_id", "kind", "slot", "ref", "created_at"}
var sessCols = []string{"id", "agent", "workspace", "title", "status", "created_at"}

func (s *Store) CreateWorkspace(ctx context.Context, ws *runtime.Workspace) error {
	if ws.ID == "" {
		ws.ID = uuid.New().String()
	}
	if ws.CreatedAt == "" {
		ws.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	query, args, _ := sq.Insert("workspaces").SetMap(map[string]any{
		"id": ws.ID, "name": ws.Name, "created_at": ws.CreatedAt,
	}).ToSql()

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *Store) GetWorkspaceByName(ctx context.Context, name string) (*runtime.Workspace, error) {
	query, args, _ := sq.Select(wsCols...).From("workspaces").
		Where(sq.Eq{"name": name}).ToSql()

	var ws runtime.Workspace
	err := s.db.GetContext(ctx, &ws, query, args...)
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func (s *Store) ListWorkspaces(ctx context.Context) ([]runtime.Workspace, error) {
	query, _, _ := sq.Select(wsCols...).From("workspaces").OrderBy("name").ToSql()

	var workspaces []runtime.Workspace
	err := s.db.SelectContext(ctx, &workspaces, query)
	return workspaces, err
}

func (s *Store) AddWorkspaceRef(ctx context.Context, ref *runtime.WorkspaceRef) error {
	if ref.CreatedAt == "" {
		ref.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	query, args, _ := sq.Insert("workspace_refs").SetMap(map[string]any{
		"workspace_id": ref.WorkspaceID, "kind": ref.Kind,
		"slot": ref.Slot, "ref": ref.Ref, "created_at": ref.CreatedAt,
	}).ToSql()

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *Store) ListWorkspaceRefs(ctx context.Context, workspaceID string) ([]runtime.WorkspaceRef, error) {
	query, args, _ := sq.Select(wsRefCols...).From("workspace_refs").
		Where(sq.Eq{"workspace_id": workspaceID}).ToSql()

	var refs []runtime.WorkspaceRef
	err := s.db.SelectContext(ctx, &refs, query, args...)
	return refs, err
}

func (s *Store) CreateSession(ctx context.Context, sess *runtime.Session) error {
	if sess.ID == "" {
		sess.ID = uuid.New().String()
	}
	if sess.Status == "" {
		sess.Status = "running"
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if sess.CreatedAt == "" {
		sess.CreatedAt = now
	}

	query, args, _ := sq.Insert("sessions").SetMap(map[string]any{
		"id": sess.ID, "agent": sess.Agent, "workspace": sess.Workspace,
		"title": sess.Title, "status": sess.Status,
		"created_at": sess.CreatedAt, "updated_at": now,
	}).ToSql()

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *Store) CompleteSession(ctx context.Context, id, status string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	query, args, _ := sq.Update("sessions").
		Set("status", status).Set("updated_at", now).
		Where(sq.Eq{"id": id}).ToSql()

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *Store) ListSessions(ctx context.Context, limit int) ([]runtime.Session, error) {
	query, args, _ := sq.Select(sessCols...).From("sessions").
		OrderBy("created_at DESC").Limit(uint64(limit)).ToSql()

	var sessions []runtime.Session
	err := s.db.SelectContext(ctx, &sessions, query, args...)
	return sessions, err
}
