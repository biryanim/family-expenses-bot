package usecase_test

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestCreateExpense_Success() {
	ctx := context.Background()
	input := entity.Expense{UserID: userA, Amount: 35000, Category: entity.CategoryFood, SpentAt: spentAt, Comment: "кофе"}

	s.repo.EXPECT().
		CreateExpense(gomock.Any(), input).
		Return(testExpense(), nil)

	result, err := s.uc.CreateExpense(ctx, input)

	s.NoError(err)
	s.Equal(testExpense(), result)
}

func (s *Suite) TestCreateExpense_DefaultSpentAt() {
	ctx := context.Background()
	input := entity.Expense{UserID: userA, Amount: 35000, Category: entity.CategoryFood}
	var saved entity.Expense

	s.repo.EXPECT().
		CreateExpense(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, e entity.Expense) (entity.Expense, error) {
			saved = e
			return testExpense(), nil
		})

	before := time.Now()
	result, err := s.uc.CreateExpense(ctx, input)
	after := time.Now()

	s.NoError(err)
	s.Equal(int64(7), result.ID)
	s.False(saved.SpentAt.Before(before))
	s.False(saved.SpentAt.After(after))
}

func (s *Suite) TestCreateExpense_MinimalAmount() {
	ctx := context.Background()
	input := entity.Expense{UserID: userA, Amount: 1, Category: entity.CategoryOther, SpentAt: spentAt}

	s.repo.EXPECT().
		CreateExpense(gomock.Any(), input).
		Return(testExpense(), nil)

	_, err := s.uc.CreateExpense(ctx, input)

	s.NoError(err)
}

func (s *Suite) TestCreateExpense_InvalidAmount() {
	for name, amount := range map[string]int64{"zero": 0, "negative": -100} {
		s.Run(name, func() {
			ctx := context.Background()

			result, err := s.uc.CreateExpense(ctx, entity.Expense{UserID: userA, Amount: amount, Category: entity.CategoryFood})

			s.Equal(entity.ErrInvalidAmount, err)
			s.Zero(result)
		})
	}
}

func (s *Suite) TestCreateExpense_InvalidCategory() {
	for name, category := range map[string]entity.Category{"unknown": "travel", "empty": ""} {
		s.Run(name, func() {
			ctx := context.Background()

			result, err := s.uc.CreateExpense(ctx, entity.Expense{UserID: userA, Amount: 35000, Category: category})

			s.Equal(entity.ErrInvalidCategory, err)
			s.Zero(result)
		})
	}
}

func (s *Suite) TestCreateExpense_UnknownUser() {
	ctx := context.Background()

	s.repo.EXPECT().
		CreateExpense(gomock.Any(), gomock.Any()).
		Return(entity.Expense{}, entity.ErrUnknownUser)

	result, err := s.uc.CreateExpense(ctx, entity.Expense{UserID: 999, Amount: 35000, Category: entity.CategoryFood, SpentAt: spentAt})

	s.ErrorIs(err, entity.ErrUnknownUser)
	s.Contains(err.Error(), "u.repo.CreateExpense")
	s.Zero(result)
}

func (s *Suite) TestCreateExpense_RepositoryError() {
	ctx := context.Background()
	repoErr := errors.New("connection refused")

	s.repo.EXPECT().
		CreateExpense(gomock.Any(), gomock.Any()).
		Return(entity.Expense{}, repoErr)

	result, err := s.uc.CreateExpense(ctx, entity.Expense{UserID: userA, Amount: 35000, Category: entity.CategoryFood, SpentAt: spentAt})

	s.Error(err)
	s.Zero(result)
	s.ErrorIs(err, repoErr)
	s.Contains(err.Error(), "u.repo.CreateExpense")
}

func (s *Suite) TestCreateExpense_CommentTooLong() {
	ctx := context.Background()
	input := testExpense()
	input.Comment = strings.Repeat("я", 201)

	result, err := s.uc.CreateExpense(ctx, input)

	s.Equal(entity.ErrInvalidComment, err)
	s.Zero(result)
}

func (s *Suite) TestCreateExpense_CommentMaxLength() {
	ctx := context.Background()
	input := testExpense()
	input.Comment = strings.Repeat("я", 200)

	s.repo.EXPECT().CreateExpense(gomock.Any(), input).Return(input, nil)

	result, err := s.uc.CreateExpense(ctx, input)

	s.NoError(err)
	s.Equal(input, result)
}

func (s *Suite) TestCreateExpense_TrimsComment() {
	ctx := context.Background()
	input := testExpense()
	input.Comment = "  " + strings.Repeat("я", 200) + "  "
	saved := testExpense()
	saved.Comment = strings.Repeat("я", 200)

	s.repo.EXPECT().CreateExpense(gomock.Any(), saved).Return(saved, nil)

	result, err := s.uc.CreateExpense(ctx, input)

	s.NoError(err)
	s.Equal(saved, result)
}
