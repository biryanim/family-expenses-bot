package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (r *Repository) UpdateExpense(ctx context.Context, e entity.Expense) (entity.Expense, error) {
	const sql = "UPDATE expenses SET amount = $1, category = $2, spent_at = $3, comment = $4, updated_at = now() WHERE id = $5 RETURNING " + expenseColumns

	var updated entity.Expense
	err := r.pool.QueryRow(ctx, sql, e.Amount, string(e.Category), e.SpentAt, e.Comment, e.ID).Scan(
		&updated.ID, &updated.UserID, &updated.Amount, &updated.Category,
		&updated.SpentAt, &updated.Comment, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Expense{}, entity.ErrNotFound
		}
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.ConstraintName {
			case constraintExpenseAmountPositive:
				return entity.Expense{}, entity.ErrInvalidAmount
			case constraintExpenseCategoryValid:
				return entity.Expense{}, entity.ErrInvalidCategory
			}
		}
		return entity.Expense{}, fmt.Errorf("r.pool.QueryRow.Scan: %w", err)
	}
	return updated, nil
}
