package skillindex

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/convexideas/oktopus/internal/store"
)

func TestIndexFilesystemAndSync(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "repo-standardization")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(`---
name: repo-standardization
description: Standardize Nava repositories. Use when repo baseline needs checking.
metadata:
  owner: platform
---
# Repo Standardization
`), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}

	src := Source{
		Name:       "local-skills",
		Type:       "filesystem",
		Scope:      "project",
		RootPath:   root,
		TrustLevel: "trusted",
		Version:    "0.1.0",
	}
	skills, err := IndexFilesystem(src)
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "repo-standardization" {
		t.Fatalf("unexpected skill name %q", skills[0].Name)
	}
	if skills[0].Path != skillPath {
		t.Fatalf("expected exact path %q, got %q", skillPath, skills[0].Path)
	}
	if skills[0].Hash == "" {
		t.Fatal("expected skill hash")
	}

	st, err := store.Open(context.Background(), "sqlite://"+filepath.Join(t.TempDir(), "oktopus.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()
	if _, err := st.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if n, err := SyncSkills(context.Background(), st.DB, src, skills); err != nil {
		t.Fatalf("sync skills: %v", err)
	} else if n != 1 {
		t.Fatalf("expected 1 synced skill, got %d", n)
	}

	var count int
	if err := st.DB.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM skill_index`).Scan(&count); err != nil {
		t.Fatalf("count skill_index: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 skill_index row, got %d", count)
	}

	if _, err := SyncSkills(context.Background(), st.DB, src, skills); err != nil {
		t.Fatalf("resync skills: %v", err)
	}
	if err := st.DB.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM skill_index`).Scan(&count); err != nil {
		t.Fatalf("recount skill_index: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected idempotent sync row count 1, got %d", count)
	}
}
