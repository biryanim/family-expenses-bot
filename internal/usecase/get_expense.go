package usecase

import (
	"context"
	"fmt"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (u *Usecase) GetExpense(ctx context.Context, id int64) (entity.Expense, error) {
	expense, err := u.repo.GetExpense(ctx, id)
	if err != nil {
		return entity.Expense{}, fmt.Errorf("u.repo.GetExpense: %w", err)
	}
	return expense, nil
}
