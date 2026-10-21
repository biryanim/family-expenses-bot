package repository

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (r *Repository) ListExpensesForUser(ctx context.Context, userID int64, f entity.ExpenseFilter) (entity.ExpensePage, error) {
	args := []any{userID}
	conditions := []string{"user_id = $1"}
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

	var where string
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	pageArgs := append(slices.Clone(args), f.Limit, f.Offset)
	sql := "SELECT " + expenseColumns + ", count(*) OVER (), (sum(amount) OVER ())::bigint FROM expenses" + where +
		" ORDER BY spent_at DESC, id DESC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)

	rows, err := r.pool.Query(ctx, sql, pageArgs...)
	if err != nil {
		return entity.ExpensePage{}, fmt.Errorf("r.pool.Query: %w", err)
	}
	defer rows.Close()

	page := entity.ExpensePage{Items: []entity.Expense{}}
	for rows.Next() {
		var e entity.Expense
		if err := rows.Scan(
			&e.ID, &e.UserID, &e.Amount, &e.Category,
			&e.SpentAt, &e.Comment, &e.CreatedAt, &e.UpdatedAt,
			&page.Count, &page.Total,
		); err != nil {
			return entity.ExpensePage{}, fmt.Errorf("rows.Scan: %w", err)
		}
		page.Items = append(page.Items, e)
	}
	if err := rows.Err(); err != nil {
		return entity.ExpensePage{}, fmt.Errorf("rows.Err: %w", err)
	}

	if len(page.Items) == 0 && f.Offset > 0 {
		const aggregatesSQL = "SELECT count(*), COALESCE(sum(amount), 0)::bigint FROM expenses"
		if err := r.pool.QueryRow(ctx, aggregatesSQL+where, args...).Scan(&page.Count, &page.Total); err != nil {
			return entity.ExpensePage{}, fmt.Errorf("r.pool.QueryRow.Scan: %w", err)
		}
	}
	return page, nil
}
