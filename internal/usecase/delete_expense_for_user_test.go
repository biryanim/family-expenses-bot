package usecase_test

import (
	"context"
	"errors"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestDeleteExpenseForUser_Success() {
	ctx := context.Background()

	s.repo.EXPECT().
		DeleteExpenseForUser(gomock.Any(), userA, int64(7)).
		Return(nil)

	err := s.uc.DeleteExpenseForUser(ctx, userA, 7)

	s.NoError(err)
}

func (s *Suite) TestDeleteExpenseForUser_NotFound() {
	ctx := context.Background()

	s.repo.EXPECT().
		DeleteExpenseForUser(gomock.Any(), userA, int64(8)).
		Return(entity.ErrNotFound)

	err := s.uc.DeleteExpenseForUser(ctx, userA, 8)

	s.ErrorIs(err, entity.ErrNotFound)
	s.Contains(err.Error(), "u.repo.DeleteExpenseForUser")
}

func (s *Suite) TestDeleteExpenseForUser_RepositoryError() {
	ctx := context.Background()
	repoErr := errors.New("connection refused")

	s.repo.EXPECT().
		DeleteExpenseForUser(gomock.Any(), userA, int64(7)).
		Return(repoErr)

	err := s.uc.DeleteExpenseForUser(ctx, userA, 7)

	s.ErrorIs(err, repoErr)
	s.Contains(err.Error(), "u.repo.DeleteExpenseForUser")
}
