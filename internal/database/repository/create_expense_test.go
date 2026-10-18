package repository_test

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/mock"
)

func newExpense() entity.Expense {
	return entity.Expense{UserID: userA, Amount: 35000, Category: entity.CategoryFood, SpentAt: spentAt, Comment: "кофе"}
}

func (s *Suite) TestCreateExpense_Success() {
	ctx := context.Background()

	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(expenseScanArgs()...).
		DoAndReturn(func(dest ...any) error {
			populateExpenseScanDest(testExpense(), dest...)
			return nil
		})

	s.pool.EXPECT().
		QueryRow(gomock.Any(), gomock.Any(), userA, int64(35000), "food", spentAt, "кофе").
		Return(mockRow)

	result, err := s.repo.CreateExpense(ctx, newExpense())

	s.NoError(err)
	s.Equal(int64(7), result.ID)
	s.Equal(userA, result.UserID)
	s.Equal(int64(35000), result.Amount)
	s.Equal(entity.CategoryFood, result.Category)
	s.Equal(spentAt, result.SpentAt)
	s.Equal("кофе", result.Comment)
	s.Equal(createdAt, result.CreatedAt)
	s.Nil(result.UpdatedAt)
}

func (s *Suite) TestCreateExpense_UnknownUser() {
	ctx := context.Background()

	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(expenseScanArgs()...).
		Return(&pgconn.PgError{Code: "23503", ConstraintName: "expenses_user_id_fkey"})

	s.pool.EXPECT().
		QueryRow(gomock.Any(), gomock.Any(), mockAnyArgs(5)...).
		Return(mockRow)

	result, err := s.repo.CreateExpense(ctx, newExpense())

	s.Error(err)
	s.Equal(entity.Expense{}, result)
	s.Equal(entity.ErrUnknownUser, err)
}

func (s *Suite) TestCreateExpense_InvalidAmount() {
	ctx := context.Background()

	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(expenseScanArgs()...).
		Return(&pgconn.PgError{Code: "23514", ConstraintName: "expenses_amount_positive"})

	s.pool.EXPECT().
		QueryRow(gomock.Any(), gomock.Any(), mockAnyArgs(5)...).
		Return(mockRow)

	result, err := s.repo.CreateExpense(ctx, newExpense())

	s.Error(err)
	s.Equal(entity.Expense{}, result)
	s.Equal(entity.ErrInvalidAmount, err)
}

func (s *Suite) TestCreateExpense_InvalidCategory() {
	ctx := context.Background()

	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(expenseScanArgs()...).
		Return(&pgconn.PgError{Code: "23514", ConstraintName: "expenses_category_valid"})

	s.pool.EXPECT().
		QueryRow(gomock.Any(), gomock.Any(), mockAnyArgs(5)...).
		Return(mockRow)

	result, err := s.repo.CreateExpense(ctx, newExpense())

	s.Error(err)
	s.Equal(entity.Expense{}, result)
	s.Equal(entity.ErrInvalidCategory, err)
}

func (s *Suite) TestCreateExpense_UnknownConstraint() {
	ctx := context.Background()
	pgErr := &pgconn.PgError{Code: "23514", ConstraintName: "some_other_check"}

	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(expenseScanArgs()...).
		Return(pgErr)

	s.pool.EXPECT().
		QueryRow(gomock.Any(), gomock.Any(), mockAnyArgs(5)...).
		Return(mockRow)

	result, err := s.repo.CreateExpense(ctx, newExpense())

	s.Error(err)
	s.Equal(entity.Expense{}, result)
	s.ErrorIs(err, pgErr)
	s.Contains(err.Error(), "r.pool.QueryRow.Scan")
}

func (s *Suite) TestCreateExpense_ScanError() {
	ctx := context.Background()
	scanErr := errors.New("scan error")

	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(expenseScanArgs()...).
		Return(scanErr)

	s.pool.EXPECT().
		QueryRow(gomock.Any(), gomock.Any(), mockAnyArgs(5)...).
		Return(mockRow)

	result, err := s.repo.CreateExpense(ctx, newExpense())

	s.Error(err)
	s.Equal(entity.Expense{}, result)
	s.ErrorIs(err, scanErr)
	s.Contains(err.Error(), "r.pool.QueryRow.Scan")
}
