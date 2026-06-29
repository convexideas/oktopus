package profile

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Profile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type LocalOptions struct{}

func InitLocal(ctx context.Context, db *sql.DB, _ LocalOptions) (Profile, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	p := Profile{
		ID:        "profile:local",
		Name:      "local",
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := db.ExecContext(ctx, `
INSERT INTO profiles (id, name, created_at, updated_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET name=excluded.name, updated_at=excluded.updated_at;`,
		p.ID, p.Name, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return Profile{}, fmt.Errorf("upsert local profile: %w", err)
	}
	return Get(ctx, db, p.ID)
}

func Get(ctx context.Context, db *sql.DB, id string) (Profile, error) {
	var p Profile
	err := db.QueryRowContext(ctx, profileSelectSQL+` WHERE id = ?`, id).Scan(scanProfile(&p)...)
	if err != nil {
		return Profile{}, fmt.Errorf("get profile %q: %w", id, err)
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
SELECT id, name, created_at, updated_at
FROM profiles`

func scanProfile(p *Profile) []any {
	return []any{&p.ID, &p.Name, &p.CreatedAt, &p.UpdatedAt}
}
