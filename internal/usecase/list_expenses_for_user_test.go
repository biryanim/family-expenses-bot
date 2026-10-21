package usecase_test

import (
	"context"
	"errors"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestListExpensesForUser_Success() {
	ctx := context.Background()
	filter := entity.ExpenseFilter{From: from, To: to, Category: entity.CategoryFood, Limit: 30, Offset: 30}

	s.repo.EXPECT().
		ListExpensesForUser(gomock.Any(), userA, filter).
		Return(testPage(), nil)

	result, err := s.uc.ListExpensesForUser(ctx, userA, filter)

	s.NoError(err)
	s.Equal(testPage(), result)
}

func (s *Suite) TestListExpensesForUser_PaginationBounds() {
	tests := []struct {
		name    string
		limit   int
		offset  int
		wantErr error
	}{
		{name: "limit min", limit: 1, offset: 0},
		{name: "limit max", limit: 500, offset: 0},
		{name: "limit zero", limit: 0, offset: 0, wantErr: entity.ErrInvalidPagination},
		{name: "limit over max", limit: 501, offset: 0, wantErr: entity.ErrInvalidPagination},
		{name: "negative limit", limit: -1, offset: 0, wantErr: entity.ErrInvalidPagination},
		{name: "negative offset", limit: 10, offset: -1, wantErr: entity.ErrInvalidPagination},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()
			filter := entity.ExpenseFilter{Limit: tt.limit, Offset: tt.offset}
			if tt.wantErr == nil {
				s.repo.EXPECT().ListExpensesForUser(gomock.Any(), userA, filter).Return(entity.ExpensePage{}, nil)
			}

			_, err := s.uc.ListExpensesForUser(ctx, userA, filter)

			if tt.wantErr == nil {
				s.NoError(err)
				return
			}
			s.Equal(tt.wantErr, err)
		})
	}
}

func (s *Suite) TestListExpensesForUser_InvalidCategory() {
	ctx := context.Background()

	result, err := s.uc.ListExpensesForUser(ctx, userA, entity.ExpenseFilter{Category: "travel", Limit: 30})

	s.Equal(entity.ErrInvalidCategory, err)
	s.Zero(result)
}

func (s *Suite) TestListExpensesForUser_InvalidPeriod() {
	ctx := context.Background()

	result, err := s.uc.ListExpensesForUser(ctx, userA, entity.ExpenseFilter{From: to, To: from, Limit: 30})

	s.Equal(entity.ErrInvalidDate, err)
	s.Zero(result)
}

func (s *Suite) TestListExpensesForUser_EqualBounds() {
	ctx := context.Background()
	filter := entity.ExpenseFilter{From: from, To: from, Limit: 30}

	s.repo.EXPECT().ListExpensesForUser(gomock.Any(), userA, filter).Return(entity.ExpensePage{Items: []entity.Expense{}}, nil)

	result, err := s.uc.ListExpensesForUser(ctx, userA, filter)

	s.NoError(err)
	s.Empty(result.Items)
}

func (s *Suite) TestListExpensesForUser_RepositoryError() {
	ctx := context.Background()
	repoErr := errors.New("connection refused")

	s.repo.EXPECT().ListExpensesForUser(gomock.Any(), userA, gomock.Any()).Return(entity.ExpensePage{}, repoErr)

	result, err := s.uc.ListExpensesForUser(ctx, userA, entity.ExpenseFilter{Limit: 30})

	s.ErrorIs(err, repoErr)
	s.Contains(err.Error(), "u.repo.ListExpensesForUser")
	s.Zero(result)
}
