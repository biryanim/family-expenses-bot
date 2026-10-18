package usecase_test

import (
	"context"
	"errors"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestGetExpense_Success() {
	ctx := context.Background()

	s.repo.EXPECT().
		GetExpense(gomock.Any(), int64(7)).
		Return(testExpense(), nil)

	result, err := s.uc.GetExpense(ctx, 7)

	s.NoError(err)
	s.Equal(testExpense(), result)
}

func (s *Suite) TestGetExpense_NotFound() {
	ctx := context.Background()

	s.repo.EXPECT().
		GetExpense(gomock.Any(), int64(999)).
		Return(entity.Expense{}, entity.ErrNotFound)

	result, err := s.uc.GetExpense(ctx, 999)

	s.ErrorIs(err, entity.ErrNotFound)
	s.Contains(err.Error(), "u.repo.GetExpense")
	s.Zero(result)
}

func (s *Suite) TestGetExpense_RepositoryError() {
	ctx := context.Background()
	repoErr := errors.New("connection refused")

	s.repo.EXPECT().
		GetExpense(gomock.Any(), int64(7)).
		Return(entity.Expense{}, repoErr)

	result, err := s.uc.GetExpense(ctx, 7)

	s.Error(err)
	s.Zero(result)
	s.ErrorIs(err, repoErr)
	s.Contains(err.Error(), "u.repo.GetExpense")
}
