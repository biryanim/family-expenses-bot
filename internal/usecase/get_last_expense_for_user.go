package usecase

import (
	"context"
	"fmt"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (u *Usecase) GetLastExpenseForUser(ctx context.Context, userID int64) (entity.Expense, error) {
	expense, err := u.repo.GetLastExpenseForUser(ctx, userID)
	if err != nil {
		return entity.Expense{}, fmt.Errorf("u.repo.GetLastExpenseForUser: %w", err)
	}
	return expense, nil
}
