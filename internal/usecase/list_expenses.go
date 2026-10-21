package usecase

import (
	"context"
	"fmt"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (u *Usecase) ListExpenses(ctx context.Context, f entity.ExpenseFilter) (entity.ExpensePage, error) {
	if f.Category != "" && !f.Category.Valid() {
		return entity.ExpensePage{}, entity.ErrInvalidCategory
	}
	if !f.From.IsZero() && !f.To.IsZero() && f.From.After(f.To) {
		return entity.ExpensePage{}, entity.ErrInvalidDate
	}
	if f.Limit < 1 || f.Limit > entity.MaxPageLimit || f.Offset < 0 {
		return entity.ExpensePage{}, entity.ErrInvalidPagination
	}

	page, err := u.repo.ListExpenses(ctx, f)
	if err != nil {
		return entity.ExpensePage{}, fmt.Errorf("u.repo.ListExpenses: %w", err)
	}
	return page, nil
}
