package repository_test

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/mock"
)

func (s *Suite) TestGetExpense_Success() {
	ctx := context.Background()
	id := int64(7)

	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(expenseScanArgs()...).
		DoAndReturn(func(dest ...any) error {
			populateExpenseScanDest(testExpense(), dest...)
			return nil
		})

	s.pool.EXPECT().
		QueryRow(gomock.Any(), selectExpenses+" WHERE id = $1", id).
		Return(mockRow)

	result, err := s.repo.GetExpense(ctx, id)

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

func (s *Suite) TestGetExpense_Updated() {
	ctx := context.Background()
	id := int64(7)
	updated := testExpense()
	updated.UpdatedAt = &updatedAt

	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(expenseScanArgs()...).
		DoAndReturn(func(dest ...any) error {
			populateExpenseScanDest(updated, dest...)
			return nil
		})

	s.pool.EXPECT().
		QueryRow(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(mockRow)

	result, err := s.repo.GetExpense(ctx, id)

	s.NoError(err)
	s.Equal(&updatedAt, result.UpdatedAt)
}

func (s *Suite) TestGetExpense_NotFound() {
	ctx := context.Background()
	id := int64(999)

	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(expenseScanArgs()...).
		Return(pgx.ErrNoRows)

	s.pool.EXPECT().
		QueryRow(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(mockRow)

	result, err := s.repo.GetExpense(ctx, id)

	s.Error(err)
	s.Equal(entity.Expense{}, result)
	s.Equal(entity.ErrNotFound, err)
}

func (s *Suite) TestGetExpense_ScanError() {
	ctx := context.Background()
	id := int64(7)
	scanErr := errors.New("scan error")

	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(expenseScanArgs()...).
		Return(scanErr)

	s.pool.EXPECT().
		QueryRow(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(mockRow)

	result, err := s.repo.GetExpense(ctx, id)

	s.Error(err)
	s.Equal(entity.Expense{}, result)
	s.ErrorIs(err, scanErr)
	s.Contains(err.Error(), "r.pool.QueryRow.Scan")
}
