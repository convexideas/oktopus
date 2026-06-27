package profile

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/convexideas/oktopus/internal/store"
)

func TestInitLocalProfileIsIdempotent(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(ctx, "sqlite://"+filepath.Join(t.TempDir(), "oktopus.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()
	if _, err := st.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	opts := LocalOptions{}
	p, err := InitLocal(ctx, st.DB, opts)
	if err != nil {
		t.Fatalf("init local: %v", err)
	}
	if p.Name != "local" || p.DefaultSandboxProvider != "openshell" {
		t.Fatalf("unexpected profile: %+v", p)
	}
	if _, err := InitLocal(ctx, st.DB, opts); err != nil {
		t.Fatalf("second init local: %v", err)
	}
	profiles, err := List(ctx, st.DB)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
}
