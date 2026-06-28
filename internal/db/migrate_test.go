package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestMigrateAppliesAndIsIdempotent(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "oktopus.db")

	conn, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer conn.Close()

	applied, err := Migrate(ctx, conn)
	if err != nil {
		t.Fatalf("first migrate: %v", err)
	}
	if len(applied) != 1 {
		t.Fatalf("expected 1 migration applied, got %d: %v", len(applied), applied)
	}

	// Second run must apply nothing (idempotent).
	applied2, err := Migrate(ctx, conn)
	if err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	if len(applied2) != 0 {
		t.Fatalf("expected 0 migrations on rerun, got %d: %v", len(applied2), applied2)
	}
}

func TestSchemaTablesExist(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "oktopus.db")
	conn, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer conn.Close()
	if _, err := Migrate(ctx, conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	tables := []string{
		"profiles", "workspaces", "workspace_refs",
		"sandbox_defs", "sandbox_instances", "sessions",
		"session_events", "session_artifacts",
		"capabilities", "skill_sources", "skill_index", "runtime_capabilities",
	}
	for _, table := range tables {
		if !hasTable(t, conn, table) {
			t.Fatalf("table %s missing", table)
		}
	}
}

func TestSchemaSessionModel(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "oktopus.db")
	conn, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer conn.Close()
	if _, err := Migrate(ctx, conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Sessions should reference sandbox_instance_id, not workspace_id directly.
	if !hasColumn(t, conn, "sessions", "sandbox_instance_id") {
		t.Fatal("sessions.sandbox_instance_id missing")
	}
	if hasColumn(t, conn, "sessions", "workspace_id") {
		t.Fatal("sessions.workspace_id should not exist (reachable via sandbox)")
	}

	// Sandbox instances reference sandbox_defs.
	if !hasColumn(t, conn, "sandbox_instances", "sandbox_def_id") {
		t.Fatal("sandbox_instances.sandbox_def_id missing")
	}
	if !hasColumn(t, conn, "sandbox_instances", "execution_principal") {
		t.Fatal("sandbox_instances.execution_principal missing")
	}

	// Sandbox defs exist with expected columns.
	for _, col := range []string{"id", "name", "provider", "base_image_ref", "created_at"} {
		if !hasColumn(t, conn, "sandbox_defs", col) {
			t.Fatalf("sandbox_defs.%s missing", col)
		}
	}
}

func hasTable(t *testing.T, conn *sql.DB, table string) bool {
	t.Helper()
	var name string
	err := conn.QueryRowContext(context.Background(),
		`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		t.Fatalf("query table %s: %v", table, err)
	}
	return true
}

func hasColumn(t *testing.T, conn *sql.DB, table, column string) bool {
	t.Helper()
	rows, err := conn.QueryContext(context.Background(), "PRAGMA table_info("+table+")")
	if err != nil {
		t.Fatalf("pragma table_info(%s): %v", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			cid       int
			name      string
			ctype     string
			notnull   int
			dfltValue sql.NullString
			pk        int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			t.Fatalf("scan column: %v", err)
		}
		if name == column {
			return true
		}
	}
	return false
}
