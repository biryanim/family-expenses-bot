package usecase

import (
	"context"
	"fmt"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (u *Usecase) ListExpenses(ctx context.Context, f entity.ExpenseFilter) ([]entity.Expense, error) {
	if f.Category != "" && !f.Category.Valid() {
		return nil, entity.ErrInvalidCategory
	}
	if !f.From.IsZero() && !f.To.IsZero() && f.From.After(f.To) {
		return nil, entity.ErrInvalidDate
	}

	expenses, err := u.repo.ListExpenses(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("u.repo.ListExpenses: %w", err)
	}
	return expenses, nil
}
