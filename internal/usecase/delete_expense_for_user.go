package usecase

import (
	"context"
	"fmt"
)

func (u *Usecase) DeleteExpenseForUser(ctx context.Context, userID, id int64) error {
	if err := u.repo.DeleteExpenseForUser(ctx, userID, id); err != nil {
		return fmt.Errorf("u.repo.DeleteExpenseForUser: %w", err)
	}
	return nil
}
