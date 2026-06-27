package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOpenSQLiteURLAndMigrate(t *testing.T) {
	ctx := context.Background()
	st, err := Open(ctx, "sqlite://"+filepath.Join(t.TempDir(), "oktopus.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer st.Close()
	if st.Engine != EngineSQLite {
		t.Fatalf("engine = %s, want %s", st.Engine, EngineSQLite)
	}
	if _, err := st.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
}

func TestPostgresAdapterIsExplicitlyUnsupported(t *testing.T) {
	if _, err := Open(context.Background(), "postgres://user:pass@localhost/oktopus"); err == nil {
		t.Fatal("expected unsupported postgres adapter error")
	}
}
