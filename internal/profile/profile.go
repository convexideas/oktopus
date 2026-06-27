package profile

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Profile struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	Scope                  string `json:"scope"`
	PolicyRef              string `json:"policy_ref,omitempty"`
	DefaultSandboxProvider string `json:"default_sandbox_provider"`
	DefaultSandboxProfile  string `json:"default_sandbox_profile"`
	DefaultAgentRef        string `json:"default_agent_ref,omitempty"`
	DefaultModelRef        string `json:"default_model_ref,omitempty"`
	BudgetRef              string `json:"budget_ref,omitempty"`
	Status                 string `json:"status"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
}

type LocalOptions struct{}

func InitLocal(ctx context.Context, db *sql.DB, _ LocalOptions) (Profile, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	p := Profile{
		ID:                     "profile:local",
		Name:                   "local",
		Scope:                  "user",
		DefaultSandboxProvider: "openshell",
		DefaultSandboxProfile:  "local-openshell",
		Status:                 "active",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	_, err := db.ExecContext(ctx, `
INSERT INTO profiles
  (id, name, scope, default_sandbox_provider, default_sandbox_profile, status, metadata_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(name) DO UPDATE SET
  scope=excluded.scope,
  default_sandbox_provider=excluded.default_sandbox_provider,
  default_sandbox_profile=excluded.default_sandbox_profile,
  status=excluded.status,
  updated_at=excluded.updated_at;`,
		p.ID, p.Name, p.Scope, p.DefaultSandboxProvider, p.DefaultSandboxProfile, p.Status, "{}", p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return Profile{}, fmt.Errorf("upsert local profile: %w", err)
	}
	return Get(ctx, db, p.Name)
}

func Get(ctx context.Context, db *sql.DB, name string) (Profile, error) {
	var p Profile
	err := db.QueryRowContext(ctx, profileSelectSQL+` WHERE name = ?`, name).Scan(scanProfile(&p)...)
	if err != nil {
		return Profile{}, fmt.Errorf("get profile %q: %w", name, err)
	}
	return p, nil
}

func List(ctx context.Context, db *sql.DB) ([]Profile, error) {
	rows, err := db.QueryContext(ctx, profileSelectSQL+` ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	defer rows.Close()

	var profiles []Profile
	for rows.Next() {
		var p Profile
		if err := rows.Scan(scanProfile(&p)...); err != nil {
			return nil, fmt.Errorf("scan profile: %w", err)
		}
		profiles = append(profiles, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate profiles: %w", err)
	}
	return profiles, nil
}

const profileSelectSQL = `
SELECT id, name, scope, COALESCE(policy_ref, ''), COALESCE(default_sandbox_provider, ''),
       COALESCE(default_sandbox_profile, ''), COALESCE(default_agent_ref, ''), COALESCE(default_model_ref, ''),
       COALESCE(budget_ref, ''), status, created_at, updated_at
FROM profiles`

func scanProfile(p *Profile) []any {
	return []any{
		&p.ID, &p.Name, &p.Scope, &p.PolicyRef, &p.DefaultSandboxProvider,
		&p.DefaultSandboxProfile, &p.DefaultAgentRef, &p.DefaultModelRef,
		&p.BudgetRef, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	}
}
