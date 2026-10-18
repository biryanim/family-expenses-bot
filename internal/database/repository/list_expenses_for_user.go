package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (r *Repository) ListExpensesForUser(ctx context.Context, userID int64, f entity.ExpenseFilter) ([]entity.Expense, error) {
	var conditions []string
	var args []any
	args = append(args, userID)
	conditions = append(conditions, "user_id = $"+strconv.Itoa(len(args)))
	if !f.From.IsZero() {
		args = append(args, f.From)
		conditions = append(conditions, "spent_at >= $"+strconv.Itoa(len(args)))
	}
	if !f.To.IsZero() {
		args = append(args, f.To)
		conditions = append(conditions, "spent_at < $"+strconv.Itoa(len(args)))
	}
	if f.Category != "" {
		args = append(args, string(f.Category))
		conditions = append(conditions, "category = $"+strconv.Itoa(len(args)))
	}

	sql := "SELECT " + expenseColumns + " FROM expenses"
	if len(conditions) > 0 {
		sql += " WHERE " + strings.Join(conditions, " AND ")
	}
	sql += " ORDER BY spent_at DESC, id DESC"

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Query: %w", err)
	}
	defer rows.Close()

	expenses := []entity.Expense{}
	for rows.Next() {
		var e entity.Expense
		if err := rows.Scan(
			&e.ID, &e.UserID, &e.Amount, &e.Category,
			&e.SpentAt, &e.Comment, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		expenses = append(expenses, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return expenses, nil
}
