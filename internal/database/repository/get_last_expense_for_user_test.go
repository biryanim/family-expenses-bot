package repository_test

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/mock"
)

func (s *Suite) TestGetLastExpenseForUser_Success() {
	ctx := context.Background()

	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(expenseScanArgs()...).
		DoAndReturn(func(dest ...any) error {
			populateExpenseScanDest(testExpense(), dest...)
			return nil
		})

	s.pool.EXPECT().
		QueryRow(gomock.Any(), selectExpenses+" WHERE user_id = $1 ORDER BY id DESC LIMIT 1", userA).
		Return(mockRow)

	result, err := s.repo.GetLastExpenseForUser(ctx, userA)

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

func (s *Suite) TestGetLastExpenseForUser_NotFound() {
	ctx := context.Background()

	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(expenseScanArgs()...).
		Return(pgx.ErrNoRows)

	s.pool.EXPECT().
		QueryRow(gomock.Any(), gomock.Any(), userB).
		Return(mockRow)

	result, err := s.repo.GetLastExpenseForUser(ctx, userB)

	s.Equal(entity.ErrNotFound, err)
	s.Equal(entity.Expense{}, result)
}

func (s *Suite) TestGetLastExpenseForUser_ScanError() {
	ctx := context.Background()
	scanErr := errors.New("scan error")

	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(expenseScanArgs()...).
		Return(scanErr)

	s.pool.EXPECT().
		QueryRow(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(mockRow)

	result, err := s.repo.GetLastExpenseForUser(ctx, userA)

	s.Error(err)
	s.Equal(entity.Expense{}, result)
	s.ErrorIs(err, scanErr)
	s.Contains(err.Error(), "r.pool.QueryRow.Scan")
}
