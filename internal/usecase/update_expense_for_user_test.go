package usecase_test

import (
	"context"
	"errors"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestUpdateExpenseForUser_Success() {
	ctx := context.Background()
	input := entity.Expense{ID: 7, Amount: 45000, Category: entity.CategoryFood, SpentAt: spentAt, Comment: "кофе"}
	updated := testExpense()
	updated.Amount = 45000
	updated.UpdatedAt = &updatedAt

	s.repo.EXPECT().
		UpdateExpenseForUser(gomock.Any(), userA, input).
		Return(updated, nil)

	result, err := s.uc.UpdateExpenseForUser(ctx, userA, input)

	s.NoError(err)
	s.Equal(updated, result)
}

func (s *Suite) TestUpdateExpenseForUser_InvalidAmount() {
	ctx := context.Background()

	result, err := s.uc.UpdateExpenseForUser(ctx, userA, entity.Expense{ID: 7, Amount: 0, Category: entity.CategoryFood, SpentAt: spentAt})

	s.Equal(entity.ErrInvalidAmount, err)
	s.Zero(result)
}

func (s *Suite) TestUpdateExpenseForUser_InvalidCategory() {
	ctx := context.Background()

	result, err := s.uc.UpdateExpenseForUser(ctx, userA, entity.Expense{ID: 7, Amount: 35000, Category: "travel", SpentAt: spentAt})

	s.Equal(entity.ErrInvalidCategory, err)
	s.Zero(result)
}

func (s *Suite) TestUpdateExpenseForUser_InvalidDate() {
	ctx := context.Background()

	result, err := s.uc.UpdateExpenseForUser(ctx, userA, entity.Expense{ID: 7, Amount: 35000, Category: entity.CategoryFood, SpentAt: time.Time{}})

	s.Equal(entity.ErrInvalidDate, err)
	s.Zero(result)
}

func (s *Suite) TestUpdateExpenseForUser_NotFound() {
	ctx := context.Background()

	s.repo.EXPECT().
		UpdateExpenseForUser(gomock.Any(), userA, gomock.Any()).
		Return(entity.Expense{}, entity.ErrNotFound)

	result, err := s.uc.UpdateExpenseForUser(ctx, userA, entity.Expense{ID: 8, Amount: 35000, Category: entity.CategoryFood, SpentAt: spentAt})

	s.ErrorIs(err, entity.ErrNotFound)
	s.Contains(err.Error(), "u.repo.UpdateExpenseForUser")
	s.Zero(result)
}

func (s *Suite) TestUpdateExpenseForUser_RepositoryError() {
	ctx := context.Background()
	repoErr := errors.New("connection refused")

	s.repo.EXPECT().
		UpdateExpenseForUser(gomock.Any(), userA, gomock.Any()).
		Return(entity.Expense{}, repoErr)

	result, err := s.uc.UpdateExpenseForUser(ctx, userA, entity.Expense{ID: 7, Amount: 35000, Category: entity.CategoryFood, SpentAt: spentAt})

	s.ErrorIs(err, repoErr)
	s.Contains(err.Error(), "u.repo.UpdateExpenseForUser")
	s.Zero(result)
}
