package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (r *Repository) UpdateExpenseForUser(ctx context.Context, userID, id int64, p entity.ExpensePatch) (entity.Expense, error) {
	const sql = "UPDATE expenses SET amount = COALESCE($1, amount), category = COALESCE($2, category), spent_at = COALESCE($3, spent_at), comment = COALESCE($4, comment), updated_at = now() WHERE id = $5 AND user_id = $6 RETURNING " + expenseColumns

	var category *string
	if p.Category != nil {
		value := string(*p.Category)
		category = &value
	}

	var updated entity.Expense
	err := r.pool.QueryRow(ctx, sql, p.Amount, category, p.SpentAt, p.Comment, id, userID).Scan(
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
