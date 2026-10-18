package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (r *Repository) GetExpenseForUser(ctx context.Context, userID, id int64) (entity.Expense, error) {
	const sql = "SELECT " + expenseColumns + " FROM expenses WHERE id = $1 AND user_id = $2"

	var e entity.Expense
	err := r.pool.QueryRow(ctx, sql, id, userID).Scan(
		&e.ID, &e.UserID, &e.Amount, &e.Category,
		&e.SpentAt, &e.Comment, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Expense{}, entity.ErrNotFound
		}
		return entity.Expense{}, fmt.Errorf("r.pool.QueryRow.Scan: %w", err)
	}
	return e, nil
}
