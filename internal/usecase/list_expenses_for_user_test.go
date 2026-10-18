package usecase_test

import (
	"context"
	"errors"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestListExpensesForUser_Success() {
	ctx := context.Background()
	filter := entity.ExpenseFilter{From: from, To: to, Category: entity.CategoryFood}

	s.repo.EXPECT().
		ListExpensesForUser(gomock.Any(), userA, filter).
		Return([]entity.Expense{testExpense()}, nil)

	result, err := s.uc.ListExpensesForUser(ctx, userA, filter)

	s.NoError(err)
	s.Equal([]entity.Expense{testExpense()}, result)
}

func (s *Suite) TestListExpensesForUser_InvalidCategory() {
	ctx := context.Background()

	result, err := s.uc.ListExpensesForUser(ctx, userA, entity.ExpenseFilter{Category: "travel"})

	s.Equal(entity.ErrInvalidCategory, err)
	s.Nil(result)
}

func (s *Suite) TestListExpensesForUser_InvalidPeriod() {
	ctx := context.Background()

	result, err := s.uc.ListExpensesForUser(ctx, userA, entity.ExpenseFilter{From: to, To: from})

	s.Equal(entity.ErrInvalidDate, err)
	s.Nil(result)
}

func (s *Suite) TestListExpensesForUser_RepositoryError() {
	ctx := context.Background()
	repoErr := errors.New("connection refused")

	s.repo.EXPECT().
		ListExpensesForUser(gomock.Any(), userA, gomock.Any()).
		Return(nil, repoErr)

	result, err := s.uc.ListExpensesForUser(ctx, userA, entity.ExpenseFilter{})

	s.ErrorIs(err, repoErr)
	s.Contains(err.Error(), "u.repo.ListExpensesForUser")
	s.Nil(result)
}
