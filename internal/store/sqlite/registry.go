package sqlite

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/convexideas/oktopus/internal/registry"
	"github.com/google/uuid"
)

var capCols = []string{"id", "kind", "name", "version", "description", "author", "source_type", "scope", "status", "manifest_json", "hash", "manifest_path", "created_at", "updated_at"}

func (s *Store) Register(ctx context.Context, c *registry.Capability) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	if c.Status == "" {
		c.Status = "active"
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if c.CreatedAt == "" {
		c.CreatedAt = now
	}
	if c.UpdatedAt == "" {
		c.UpdatedAt = now
	}

	query, args, _ := sq.Insert("capabilities").SetMap(map[string]any{
		"id": c.ID, "kind": c.Kind, "name": c.Name, "version": c.Version,
		"description": c.Description, "author": c.Author, "source_type": c.SourceType,
		"scope": c.Scope, "status": c.Status, "manifest_json": c.ManifestJSON,
		"hash": c.Hash, "manifest_path": c.ManifestPath,
		"created_at": c.CreatedAt, "updated_at": c.UpdatedAt,
	}).ToSql()

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *Store) FindCapability(ctx context.Context, kind, name, version string) (*registry.Capability, error) {
	query, args, _ := sq.Select(capCols...).From("capabilities").
		Where(sq.Eq{"kind": kind, "name": name, "version": version, "status": "active"}).
		Limit(1).ToSql()

	var c registry.Capability
	err := s.db.GetContext(ctx, &c, query, args...)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) FindCapabilityByName(ctx context.Context, kind, name string) (*registry.Capability, error) {
	query, args, _ := sq.Select(capCols...).From("capabilities").
		Where(sq.Eq{"kind": kind, "name": name, "status": "active"}).
		OrderBy("created_at DESC").Limit(1).ToSql()

	var c registry.Capability
	err := s.db.GetContext(ctx, &c, query, args...)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) ListCapabilitiesByKind(ctx context.Context, kind string) ([]registry.Capability, error) {
	query, args, _ := sq.Select(capCols...).From("capabilities").
		Where(sq.Eq{"kind": kind, "status": "active"}).
		OrderBy("name", "version").ToSql()

	var caps []registry.Capability
	err := s.db.SelectContext(ctx, &caps, query, args...)
	return caps, err
}

func (s *Store) ListCapabilities(ctx context.Context) ([]registry.Capability, error) {
	query, args, _ := sq.Select(capCols...).From("capabilities").
		Where(sq.Eq{"status": "active"}).
		OrderBy("kind", "name", "version").ToSql()

	var caps []registry.Capability
	err := s.db.SelectContext(ctx, &caps, query, args...)
	return caps, err
}
