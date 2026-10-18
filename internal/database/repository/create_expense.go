package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (r *Repository) CreateExpense(ctx context.Context, e entity.Expense) (entity.Expense, error) {
	const sql = "INSERT INTO expenses (user_id, amount, category, spent_at, comment) VALUES ($1, $2, $3, $4, $5) RETURNING " + expenseColumns

	var created entity.Expense
	err := r.pool.QueryRow(ctx, sql, e.UserID, e.Amount, string(e.Category), e.SpentAt, e.Comment).Scan(
		&created.ID, &created.UserID, &created.Amount, &created.Category,
		&created.SpentAt, &created.Comment, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.ConstraintName {
			case constraintExpenseUserFK:
				return entity.Expense{}, entity.ErrUnknownUser
			case constraintExpenseAmountPositive:
				return entity.Expense{}, entity.ErrInvalidAmount
			case constraintExpenseCategoryValid:
				return entity.Expense{}, entity.ErrInvalidCategory
			}
		}
		return entity.Expense{}, fmt.Errorf("r.pool.QueryRow.Scan: %w", err)
	}
	return created, nil
}
