package registry

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/convexideas/oktopus/internal/store"
)

// TestSyncCapabilities loads the repo's seed registry, syncs it into a fresh
// relational index, and verifies the capabilities table is populated and that a
// re-sync is idempotent (upsert, not duplicate-insert).
func TestSyncCapabilities(t *testing.T) {
	ctx := context.Background()

	// Repo registry lives two levels up from internal/registry.
	regPath := filepath.Join("..", "..", "registry")
	reg, err := Load(regPath)
	if err != nil {
		t.Fatalf("load registry %s: %v", regPath, err)
	}
	if len(reg.Capabilities) == 0 {
		t.Fatal("expected seed registry to contain capabilities")
	}

	st, err := store.Open(ctx, "sqlite://"+filepath.Join(t.TempDir(), "oktopus.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()
	if _, err := st.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	n, err := SyncCapabilities(ctx, st.DB, reg)
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if n != len(reg.Capabilities) {
		t.Fatalf("expected %d synced, got %d", len(reg.Capabilities), n)
	}

	var count int
	if err := st.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM capabilities").Scan(&count); err != nil {
		t.Fatalf("count capabilities: %v", err)
	}
	if count != len(reg.Capabilities) {
		t.Fatalf("expected %d rows in capabilities, got %d", len(reg.Capabilities), count)
	}

	// Re-sync must not create duplicates (ON CONFLICT upsert).
	if _, err := SyncCapabilities(ctx, st.DB, reg); err != nil {
		t.Fatalf("resync: %v", err)
	}
	var count2 int
	if err := st.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM capabilities").Scan(&count2); err != nil {
		t.Fatalf("recount capabilities: %v", err)
	}
	if count2 != count {
		t.Fatalf("resync changed row count: was %d, now %d", count, count2)
	}
}
