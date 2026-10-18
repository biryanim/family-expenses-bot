package repository_test

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/mock"
)

const updateExpenseSQL = "UPDATE expenses SET amount = $1, category = $2, spent_at = $3, comment = $4, updated_at = now() WHERE id = $5 RETURNING id, user_id, amount, category, spent_at, comment, created_at, updated_at"

func (s *Suite) TestUpdateExpense_Success() {
	ctx := context.Background()
	input := testExpense()
	input.Amount = 120000
	updated := testExpense()
	updated.Amount = 120000
	updated.UpdatedAt = &updatedAt

	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(expenseScanArgs()...).
		DoAndReturn(func(dest ...any) error {
			populateExpenseScanDest(updated, dest...)
			return nil
		})

	s.pool.EXPECT().
		QueryRow(gomock.Any(), updateExpenseSQL, int64(120000), "food", spentAt, "кофе", int64(7)).
		Return(mockRow)

	result, err := s.repo.UpdateExpense(ctx, input)

	s.NoError(err)
	s.Equal(int64(7), result.ID)
	s.Equal(userA, result.UserID)
	s.Equal(int64(120000), result.Amount)
	s.Equal(entity.CategoryFood, result.Category)
	s.Equal(spentAt, result.SpentAt)
	s.Equal("кофе", result.Comment)
	s.Equal(createdAt, result.CreatedAt)
	s.Equal(&updatedAt, result.UpdatedAt)
}

func (s *Suite) TestUpdateExpense_5OTFOUND() {
	ctx := context.Background()
	input := testExpense()

	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(expenseScanArgs()...).
		Return(pgx.ErrNoRows)

	s.pool.EXPECT().
		QueryRow(gomock.Any(), updateExpenseSQL, mockAnyArgs(5)...).
		Return(mockRow)

	result, err := s.repo.UpdateExpense(ctx, input)

	s.Error(err)
	s.Equal(entity.Expense{}, result)
	s.Equal(entity.ErrNotFound, err)
}

func (s *Suite) TestUpdateExpense_InvalidAmount() {
	ctx := context.Background()
	input := testExpense()
	input.Amount = 0

	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(expenseScanArgs()...).
		Return(&pgconn.PgError{Code: "23514", ConstraintName: "expenses_amount_positive"})

	s.pool.EXPECT().
		QueryRow(gomock.Any(), gomock.Any(), mockAnyArgs(5)...).
		Return(mockRow)

	result, err := s.repo.UpdateExpense(ctx, input)

	s.Error(err)
	s.Equal(entity.Expense{}, result)
	s.Equal(entity.ErrInvalidAmount, err)
}

func (s *Suite) TestUpdateExpense_InvalidCategory() {
	ctx := context.Background()
	input := testExpense()
	input.Category = "travel"

	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(expenseScanArgs()...).
		Return(&pgconn.PgError{Code: "23514", ConstraintName: "expenses_category_valid"})

	s.pool.EXPECT().
		QueryRow(gomock.Any(), gomock.Any(), mockAnyArgs(5)...).
		Return(mockRow)

	result, err := s.repo.UpdateExpense(ctx, input)

	s.Error(err)
	s.Equal(entity.Expense{}, result)
	s.Equal(entity.ErrInvalidCategory, err)
}

func (s *Suite) TestUpdateExpense_ScanError() {
	ctx := context.Background()
	input := testExpense()
	scanErr := errors.New("scan error")

	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(expenseScanArgs()...).
		Return(scanErr)

	s.pool.EXPECT().
		QueryRow(gomock.Any(), gomock.Any(), mockAnyArgs(5)...).
		Return(mockRow)

	result, err := s.repo.UpdateExpense(ctx, input)

	s.Error(err)
	s.Equal(entity.Expense{}, result)
	s.ErrorIs(err, scanErr)
	s.Contains(err.Error(), "r.pool.QueryRow.Scan")
}
