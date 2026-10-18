package usecase

import (
	"context"
	"fmt"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (u *Usecase) ListExpensesForUser(ctx context.Context, userID int64, f entity.ExpenseFilter) ([]entity.Expense, error) {
	if f.Category != "" && !f.Category.Valid() {
		return nil, entity.ErrInvalidCategory
	}
	if !f.From.IsZero() && !f.To.IsZero() && f.From.After(f.To) {
		return nil, entity.ErrInvalidDate
	}

	expenses, err := u.repo.ListExpensesForUser(ctx, userID, f)
	if err != nil {
		return nil, fmt.Errorf("u.repo.ListExpensesForUser: %w", err)
	}
	return expenses, nil
}
