package usecase

import (
	"context"
	"fmt"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (u *Usecase) UpdateExpense(ctx context.Context, e entity.Expense) (entity.Expense, error) {
	if e.Amount <= 0 {
		return entity.Expense{}, entity.ErrInvalidAmount
	}
	if !e.Category.Valid() {
		return entity.Expense{}, entity.ErrInvalidCategory
	}
	if e.SpentAt.IsZero() {
		return entity.Expense{}, entity.ErrInvalidDate
	}

	updated, err := u.repo.UpdateExpense(ctx, e)
	if err != nil {
		return entity.Expense{}, fmt.Errorf("u.repo.UpdateExpense: %w", err)
	}
	return updated, nil
}
