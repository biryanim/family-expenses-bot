package usecase_test

import (
	"context"
	"errors"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestUpdateExpense_Success() {
	ctx := context.Background()
	input := entity.Expense{ID: 7, Amount: 150000, Category: entity.CategoryTransport, SpentAt: spentAt}
	updated := testExpense()
	updated.Amount = 150000
	updated.Category = entity.CategoryTransport
	updated.Comment = ""
	updated.UpdatedAt = &updatedAt

	s.repo.EXPECT().
		UpdateExpense(gomock.Any(), input).
		Return(updated, nil)

	result, err := s.uc.UpdateExpense(ctx, input)

	s.NoError(err)
	s.Equal(updated, result)
}

func (s *Suite) TestUpdateExpense_InvalidAmount() {
	ctx := context.Background()

	result, err := s.uc.UpdateExpense(ctx, entity.Expense{ID: 7, Amount: 0, Category: entity.CategoryFood, SpentAt: spentAt})

	s.Equal(entity.ErrInvalidAmount, err)
	s.Zero(result)
}

func (s *Suite) TestUpdateExpense_InvalidCategory() {
	ctx := context.Background()

	result, err := s.uc.UpdateExpense(ctx, entity.Expense{ID: 7, Amount: 35000, Category: "travel", SpentAt: spentAt})

	s.Equal(entity.ErrInvalidCategory, err)
	s.Zero(result)
}

func (s *Suite) TestUpdateExpense_InvalidDate() {
	ctx := context.Background()

	result, err := s.uc.UpdateExpense(ctx, entity.Expense{ID: 7, Amount: 35000, Category: entity.CategoryFood, SpentAt: time.Time{}})

	s.Equal(entity.ErrInvalidDate, err)
	s.Zero(result)
}

func (s *Suite) TestUpdateExpense_NotFound() {
	ctx := context.Background()

	s.repo.EXPECT().
		UpdateExpense(gomock.Any(), gomock.Any()).
		Return(entity.Expense{}, entity.ErrNotFound)

	result, err := s.uc.UpdateExpense(ctx, entity.Expense{ID: 999, Amount: 35000, Category: entity.CategoryFood, SpentAt: spentAt})

	s.ErrorIs(err, entity.ErrNotFound)
	s.Contains(err.Error(), "u.repo.UpdateExpense")
	s.Zero(result)
}

func (s *Suite) TestUpdateExpense_RepositoryError() {
	ctx := context.Background()
	repoErr := errors.New("connection refused")

	s.repo.EXPECT().
		UpdateExpense(gomock.Any(), gomock.Any()).
		Return(entity.Expense{}, repoErr)

	result, err := s.uc.UpdateExpense(ctx, entity.Expense{ID: 7, Amount: 35000, Category: entity.CategoryFood, SpentAt: spentAt})

	s.Error(err)
	s.Zero(result)
	s.ErrorIs(err, repoErr)
	s.Contains(err.Error(), "u.repo.UpdateExpense")
}
