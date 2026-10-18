package repository

import (
	"context"
	"fmt"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (r *Repository) DeleteExpense(ctx context.Context, id int64) error {
	const sql = "DELETE FROM expenses WHERE id = $1"

	tag, err := r.pool.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return entity.ErrNotFound
	}
	return nil
}
