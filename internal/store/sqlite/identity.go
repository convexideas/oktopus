package sqlite

import (
	"context"
	"encoding/json"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/convexideas/oktopus/internal/identity"
	"github.com/google/uuid"
)

var profileCols = []string{"id", "user_id", "config_json", "created_at", "updated_at"}

func (s *Store) GetProfile(ctx context.Context, userID string) (*identity.Profile, error) {
	query, args, _ := sq.Select(profileCols...).From("profiles").
		Where(sq.Eq{"user_id": userID}).ToSql()

	var p identity.Profile
	if err := s.db.GetContext(ctx, &p, query, args...); err != nil {
		return nil, err
	}

	// Hydrate from JSON
	if p.DataJSON != "" {
		var data profileData
		json.Unmarshal([]byte(p.DataJSON), &data)
		p.Preferences = data.Preferences
		p.HarnessConfig = data.HarnessConfig
	}

	return &p, nil
}

func (s *Store) UpsertProfile(ctx context.Context, p *identity.Profile) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if p.CreatedAt == "" {
		p.CreatedAt = now
	}
	p.UpdatedAt = now

	// Serialize to JSON
	data := profileData{
		Preferences:   p.Preferences,
		HarnessConfig: p.HarnessConfig,
	}
	dataJSON, _ := json.Marshal(data)
	p.DataJSON = string(dataJSON)

	query, args, _ := sq.Insert("profiles").SetMap(map[string]any{
		"id": p.ID, "user_id": p.UserID,
		"default_model":   p.Preferences.DefaultModel,
		"default_runtime": p.Preferences.DefaultRuntime,
		"config_json":     p.DataJSON,
		"created_at":      p.CreatedAt, "updated_at": p.UpdatedAt,
	}).Suffix(`ON CONFLICT(user_id) DO UPDATE SET
		default_model = excluded.default_model,
		default_runtime = excluded.default_runtime,
		config_json = excluded.config_json,
		updated_at = excluded.updated_at`).ToSql()

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

type profileData struct {
	Preferences   identity.Preferences   `json:"preferences"`
	HarnessConfig identity.HarnessConfig `json:"harness_config"`
}
