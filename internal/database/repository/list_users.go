package repository

import (
	"context"
	"fmt"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (r *Repository) ListUsers(ctx context.Context) ([]entity.User, error) {
	const sql = "SELECT telegram_id, username, first_name, last_name, created_at, updated_at FROM users ORDER BY telegram_id"

	rows, err := r.pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Query: %w", err)
	}
	defer rows.Close()

	users := []entity.User{}
	for rows.Next() {
		var u entity.User
		if err := rows.Scan(&u.TelegramID, &u.Username, &u.FirstName, &u.LastName, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return users, nil
}
