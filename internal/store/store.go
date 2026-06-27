package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/convexideas/oktopus/internal/db"
)

type Engine string

const (
	EngineSQLite Engine = "sqlite"
)

func DefaultURL() string {
	return "sqlite://.oktopus/oktopus.db"
}

type Store struct {
	Engine Engine
	URL    string
	Path   string
	DB     *sql.DB
}

func Open(ctx context.Context, databaseURL string) (*Store, error) {
	_ = ctx // reserved for remote adapters that need connection negotiation.
	if databaseURL == "" {
		databaseURL = DefaultURL()
	}

	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		return nil, fmt.Errorf("postgres store adapter not implemented yet")
	}

	path := databaseURL
	if strings.HasPrefix(databaseURL, "sqlite://") {
		path = strings.TrimPrefix(databaseURL, "sqlite://")
	}
	if path == "" {
		return nil, fmt.Errorf("sqlite database path is empty")
	}

	conn, err := db.Open(path)
	if err != nil {
		return nil, err
	}
	return &Store{Engine: EngineSQLite, URL: databaseURL, Path: path, DB: conn}, nil
}

func (s *Store) Close() error {
	if s == nil || s.DB == nil {
		return nil
	}
	return s.DB.Close()
}

func (s *Store) Migrate(ctx context.Context) ([]string, error) {
	if s == nil || s.DB == nil {
		return nil, fmt.Errorf("store is not open")
	}
	return db.Migrate(ctx, s.DB)
}

func (s *Store) Location() string {
	if s == nil {
		return ""
	}
	if s.Path != "" {
		return s.Path
	}
	return s.URL
}

func LocalPath(databaseURL string) (string, bool) {
	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		return "", false
	}
	return strings.TrimPrefix(databaseURL, "sqlite://"), true
}
