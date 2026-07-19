package sqlite

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/convexideas/oktopus/internal/gateway"
)

func (s *Store) SaveCaptures(ctx context.Context, captures []gateway.Capture) error {
	for _, c := range captures {
		query, args, _ := sq.Insert("gateway_captures").SetMap(map[string]any{
			"id":            c.ID,
			"session_id":    c.SessionID,
			"timestamp":     c.Timestamp.Format("2006-01-02T15:04:05Z"),
			"provider":      c.Provider,
			"model":         c.Model,
			"endpoint":      c.Endpoint,
			"tokens_in":     c.TokensIn,
			"tokens_out":    c.TokensOut,
			"cost_usd":      c.Cost,
			"duration_ms":   c.Duration.Milliseconds(),
			"request_json":  string(c.Request),
			"response_json": string(c.Response),
		}).ToSql()
		if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}
	return nil
}
