package store

import (
	"context"
	"fmt"
)

// GetConfig retrieves a value from the config table by key.
func (s *PgStore) GetConfig(ctx context.Context, key string) (string, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	var value string
	err := s.pool.QueryRow(ctx, "SELECT value FROM config WHERE key = $1", key).Scan(&value)
	if err != nil {
		return "", fmt.Errorf("config key %q: %w", key, err)
	}
	return value, nil
}

// SetConfig upserts a value in the config table.
func (s *PgStore) SetConfig(ctx context.Context, key, value string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	_, err := s.pool.Exec(ctx,
		`INSERT INTO config (key, value) VALUES ($1, $2)
		 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`,
		key, value)
	if err != nil {
		return fmt.Errorf("setting config %q: %w", key, err)
	}
	return nil
}
