package sqlite

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/convexideas/oktopus/internal/memory"
	"github.com/google/uuid"
)

var episodeCols = []string{"id", "session_id", "workspace_id", "source", "content", "captured_at"}

func (s *Store) SaveEpisode(ctx context.Context, ep *memory.Episode) error {
	if ep.ID == "" {
		ep.ID = uuid.New().String()
	}
	if ep.CapturedAt == "" {
		ep.CapturedAt = time.Now().UTC().Format(time.RFC3339)
	}

	query, args, _ := sq.Insert("episodes").SetMap(map[string]any{
		"id": ep.ID, "session_id": ep.SessionID, "workspace_id": ep.WorkspaceID,
		"source": ep.Source, "content": ep.Content, "captured_at": ep.CapturedAt,
	}).ToSql()

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *Store) ListEpisodesByWorkspace(ctx context.Context, workspaceID string, limit int) ([]memory.Episode, error) {
	query, args, _ := sq.Select(episodeCols...).From("episodes").
		Where(sq.Eq{"workspace_id": workspaceID}).
		OrderBy("captured_at DESC").Limit(uint64(limit)).ToSql()

	var eps []memory.Episode
	err := s.db.SelectContext(ctx, &eps, query, args...)
	return eps, err
}

func (s *Store) ListSummariesByWorkspace(ctx context.Context, workspaceID string, limit int) ([]memory.Episode, error) {
	query, args, _ := sq.Select(episodeCols...).From("episodes").
		Where(sq.Eq{"workspace_id": workspaceID, "source": "summary"}).
		OrderBy("captured_at DESC").Limit(uint64(limit)).ToSql()

	var eps []memory.Episode
	err := s.db.SelectContext(ctx, &eps, query, args...)
	return eps, err
}

func (s *Store) ListRecentEpisodes(ctx context.Context, limit int) ([]memory.Episode, error) {
	query, args, _ := sq.Select(episodeCols...).From("episodes").
		OrderBy("captured_at DESC").Limit(uint64(limit)).ToSql()

	var eps []memory.Episode
	err := s.db.SelectContext(ctx, &eps, query, args...)
	return eps, err
}
