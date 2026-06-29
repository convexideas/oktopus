package db

import (
	"context"
	"path/filepath"
	"testing"
)

func TestMigrateAppliesSchemaOnce(t *testing.T) {
	ctx := context.Background()
	db, err := Open(filepath.Join(t.TempDir(), "oktopus.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	applied, err := Migrate(ctx, db)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if len(applied) != 1 || applied[0] != "0001_schema" {
		t.Fatalf("applied = %#v, want [0001_schema]", applied)
	}

	for _, table := range []string{
		"profiles",
		"workspaces",
		"workspace_refs",
		"sandbox_defs",
		"sandbox_instances",
		"sessions",
		"messages",
		"message_parts",
		"message_attachments",
		"capabilities",
	} {
		var name string
		err := db.QueryRowContext(ctx, `SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&name)
		if err != nil {
			t.Fatalf("table %s missing: %v", table, err)
		}
	}

	applied, err = Migrate(ctx, db)
	if err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	if len(applied) != 0 {
		t.Fatalf("second applied = %#v, want none", applied)
	}
}
