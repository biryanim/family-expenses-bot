package usecase

import (
	"context"
	"fmt"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (u *Usecase) RegisterMember(ctx context.Context, user entity.User) error {
	if err := u.repo.UpsertUser(ctx, user); err != nil {
		return fmt.Errorf("u.repo.UpsertUser: %w", err)
	}
	return nil
}
