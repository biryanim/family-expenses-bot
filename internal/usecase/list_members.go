package usecase

import (
	"context"
	"fmt"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (u *Usecase) ListMembers(ctx context.Context) ([]entity.User, error) {
	users, err := u.repo.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("u.repo.ListUsers: %w", err)
	}
	return users, nil
}
