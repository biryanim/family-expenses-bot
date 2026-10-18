package usecase_test

import (
	"context"
	"errors"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestGetExpenseForUser_Success() {
	ctx := context.Background()

	s.repo.EXPECT().
		GetExpenseForUser(gomock.Any(), userA, int64(7)).
		Return(testExpense(), nil)

	result, err := s.uc.GetExpenseForUser(ctx, userA, 7)

	s.NoError(err)
	s.Equal(testExpense(), result)
}

func (s *Suite) TestGetExpenseForUser_NotFound() {
	ctx := context.Background()

	s.repo.EXPECT().
		GetExpenseForUser(gomock.Any(), userA, int64(8)).
		Return(entity.Expense{}, entity.ErrNotFound)

	result, err := s.uc.GetExpenseForUser(ctx, userA, 8)

	s.ErrorIs(err, entity.ErrNotFound)
	s.Contains(err.Error(), "u.repo.GetExpenseForUser")
	s.Zero(result)
}

func (s *Suite) TestGetExpenseForUser_RepositoryError() {
	ctx := context.Background()
	repoErr := errors.New("connection refused")

	s.repo.EXPECT().
		GetExpenseForUser(gomock.Any(), userA, int64(7)).
		Return(entity.Expense{}, repoErr)

	result, err := s.uc.GetExpenseForUser(ctx, userA, 7)

	s.ErrorIs(err, repoErr)
	s.Contains(err.Error(), "u.repo.GetExpenseForUser")
	s.Zero(result)
}
