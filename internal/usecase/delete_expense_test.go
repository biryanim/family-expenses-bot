package usecase_test

import (
	"context"
	"errors"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestDeleteExpense_Success() {
	ctx := context.Background()

	s.repo.EXPECT().
		DeleteExpense(gomock.Any(), int64(7)).
		Return(nil)

	err := s.uc.DeleteExpense(ctx, 7)

	s.NoError(err)
}

func (s *Suite) TestDeleteExpense_NotFound() {
	ctx := context.Background()

	s.repo.EXPECT().
		DeleteExpense(gomock.Any(), int64(999)).
		Return(entity.ErrNotFound)

	err := s.uc.DeleteExpense(ctx, 999)

	s.ErrorIs(err, entity.ErrNotFound)
	s.Contains(err.Error(), "u.repo.DeleteExpense")
}

func (s *Suite) TestDeleteExpense_RepositoryError() {
	ctx := context.Background()
	repoErr := errors.New("connection refused")

	s.repo.EXPECT().
		DeleteExpense(gomock.Any(), int64(7)).
		Return(repoErr)

	err := s.uc.DeleteExpense(ctx, 7)

	s.Error(err)
	s.ErrorIs(err, repoErr)
	s.Contains(err.Error(), "u.repo.DeleteExpense")
}
