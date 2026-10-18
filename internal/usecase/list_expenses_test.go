package usecase_test

import (
	"context"
	"errors"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestListExpenses_Success() {
	ctx := context.Background()
	user := userA
	filter := entity.ExpenseFilter{UserID: &user, From: from, To: to, Category: entity.CategoryFood}

	s.repo.EXPECT().
		ListExpenses(gomock.Any(), filter).
		Return([]entity.Expense{testExpense()}, nil)

	result, err := s.uc.ListExpenses(ctx, filter)

	s.NoError(err)
	s.Equal([]entity.Expense{testExpense()}, result)
}

func (s *Suite) TestListExpenses_NoFilter() {
	ctx := context.Background()

	s.repo.EXPECT().
		ListExpenses(gomock.Any(), entity.ExpenseFilter{}).
		Return([]entity.Expense{}, nil)

	result, err := s.uc.ListExpenses(ctx, entity.ExpenseFilter{})

	s.NoError(err)
	s.NotNil(result)
	s.Empty(result)
}

func (s *Suite) TestListExpenses_EqualBounds() {
	ctx := context.Background()
	filter := entity.ExpenseFilter{From: from, To: from}

	s.repo.EXPECT().
		ListExpenses(gomock.Any(), filter).
		Return([]entity.Expense{}, nil)

	result, err := s.uc.ListExpenses(ctx, filter)

	s.NoError(err)
	s.Empty(result)
}

func (s *Suite) TestListExpenses_InvalidCategory() {
	ctx := context.Background()

	result, err := s.uc.ListExpenses(ctx, entity.ExpenseFilter{Category: "travel"})

	s.Equal(entity.ErrInvalidCategory, err)
	s.Nil(result)
}

func (s *Suite) TestListExpenses_InvalidPeriod() {
	ctx := context.Background()

	result, err := s.uc.ListExpenses(ctx, entity.ExpenseFilter{From: to, To: from})

	s.Equal(entity.ErrInvalidDate, err)
	s.Nil(result)
}

func (s *Suite) TestListExpenses_RepositoryError() {
	ctx := context.Background()
	repoErr := errors.New("connection refused")

	s.repo.EXPECT().
		ListExpenses(gomock.Any(), gomock.Any()).
		Return(nil, repoErr)

	result, err := s.uc.ListExpenses(ctx, entity.ExpenseFilter{})

	s.Error(err)
	s.Nil(result)
	s.ErrorIs(err, repoErr)
	s.Contains(err.Error(), "u.repo.ListExpenses")
}
