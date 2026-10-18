package usecase

import (
	"context"
	"fmt"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (u *Usecase) GetExpenseForUser(ctx context.Context, userID, id int64) (entity.Expense, error) {
	expense, err := u.repo.GetExpenseForUser(ctx, userID, id)
	if err != nil {
		return entity.Expense{}, fmt.Errorf("u.repo.GetExpenseForUser: %w", err)
	}
	return expense, nil
}
