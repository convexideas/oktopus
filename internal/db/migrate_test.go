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
		"users", "workspaces", "workspace_refs",
		"sandbox_defs", "sandbox_instances", "sessions",
		"messages", "message_parts", "message_attachments",
		"capabilities",
	}
	for _, table := range tables {
		if !hasTable(t, conn, table) {
			t.Fatalf("table %s missing", table)
		}
	}

	// Old tables should not exist.
	for _, table := range []string{"skill_sources", "skill_index", "runtime_capabilities"} {
		if hasTable(t, conn, table) {
			t.Fatalf("%s should not exist in flattened schema", table)
		}
	}
}

func TestSchemaMessageModel(t *testing.T) {
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

	// Messages have parent_id for lineage.
	for _, col := range []string{"id", "session_id", "parent_id", "role", "sequence", "created_at"} {
		if !hasColumn(t, conn, "messages", col) {
			t.Fatalf("messages.%s missing", col)
		}
	}

	// Message parts have type + content_json.
	for _, col := range []string{"id", "message_id", "sequence", "type", "content_json"} {
		if !hasColumn(t, conn, "message_parts", col) {
			t.Fatalf("message_parts.%s missing", col)
		}
	}

	// Message attachments have direction (input/output).
	for _, col := range []string{"id", "message_id", "direction", "name", "uri"} {
		if !hasColumn(t, conn, "message_attachments", col) {
			t.Fatalf("message_attachments.%s missing", col)
		}
	}

	// Old tables should not exist.
	for _, table := range []string{"session_events", "session_artifacts"} {
		if hasTable(t, conn, table) {
			t.Fatalf("%s should not exist in new schema", table)
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
