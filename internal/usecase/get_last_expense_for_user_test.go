package usecase_test

import (
	"context"
	"errors"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestGetLastExpenseForUser_Success() {
	ctx := context.Background()

	s.repo.EXPECT().
		GetLastExpenseForUser(gomock.Any(), userA).
		Return(testExpense(), nil)

	result, err := s.uc.GetLastExpenseForUser(ctx, userA)

	s.NoError(err)
	s.Equal(testExpense(), result)
}

func (s *Suite) TestGetLastExpenseForUser_NotFound() {
	ctx := context.Background()

	s.repo.EXPECT().
		GetLastExpenseForUser(gomock.Any(), userA).
		Return(entity.Expense{}, entity.ErrNotFound)

	result, err := s.uc.GetLastExpenseForUser(ctx, userA)

	s.ErrorIs(err, entity.ErrNotFound)
	s.Contains(err.Error(), "u.repo.GetLastExpenseForUser")
	s.Zero(result)
}

func (s *Suite) TestGetLastExpenseForUser_RepositoryError() {
	ctx := context.Background()
	repoErr := errors.New("connection refused")

	s.repo.EXPECT().
		GetLastExpenseForUser(gomock.Any(), userA).
		Return(entity.Expense{}, repoErr)

	result, err := s.uc.GetLastExpenseForUser(ctx, userA)

	s.ErrorIs(err, repoErr)
	s.Contains(err.Error(), "u.repo.GetLastExpenseForUser")
	s.Zero(result)
}
