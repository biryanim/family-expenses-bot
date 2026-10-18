package usecase

import (
	"context"
	"fmt"
)

func (u *Usecase) DeleteExpense(ctx context.Context, id int64) error {
	if err := u.repo.DeleteExpense(ctx, id); err != nil {
		return fmt.Errorf("u.repo.DeleteExpense: %w", err)
	}
	return nil
}
