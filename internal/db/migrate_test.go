package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestMigrateAppliesAllAndIsIdempotent(t *testing.T) {
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
	if len(applied) < 2 {
		t.Fatalf("expected at least 2 migrations applied, got %d: %v", len(applied), applied)
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

func TestMigration0002Schema(t *testing.T) {
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

	// project_counters table exists.
	var name string
	if err := conn.QueryRowContext(ctx,
		`SELECT name FROM sqlite_master WHERE type='table' AND name='project_counters'`).Scan(&name); err != nil {
		t.Fatalf("project_counters table missing: %v", err)
	}

	// jobs.retry_policy_json column exists.
	if !hasColumn(t, conn, "jobs", "retry_policy_json") {
		t.Fatal("jobs.retry_policy_json column missing")
	}
	// approvals.attempt_id column exists.
	if !hasColumn(t, conn, "approvals", "attempt_id") {
		t.Fatal("approvals.attempt_id column missing")
	}
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
