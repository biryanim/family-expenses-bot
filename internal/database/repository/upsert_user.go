package repository

import (
	"context"
	"fmt"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (r *Repository) UpsertUser(ctx context.Context, u entity.User) error {
	const sql = "INSERT INTO users (telegram_id, username, first_name, last_name) VALUES ($1, $2, $3, $4)" +
		" ON CONFLICT (telegram_id) DO UPDATE SET username = EXCLUDED.username, first_name = EXCLUDED.first_name," +
		" last_name = EXCLUDED.last_name, updated_at = now()" +
		" WHERE (users.username, users.first_name, users.last_name)" +
		" IS DISTINCT FROM (EXCLUDED.username, EXCLUDED.first_name, EXCLUDED.last_name)"

	if _, err := r.pool.Exec(ctx, sql, u.TelegramID, u.Username, u.FirstName, u.LastName); err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}
