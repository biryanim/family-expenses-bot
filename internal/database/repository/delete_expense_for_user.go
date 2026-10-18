package repository

import (
	"context"
	"fmt"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (r *Repository) DeleteExpenseForUser(ctx context.Context, userID, id int64) error {
	const sql = "DELETE FROM expenses WHERE id = $1 AND user_id = $2"

	tag, err := r.pool.Exec(ctx, sql, id, userID)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return entity.ErrNotFound
	}
	return nil
}
