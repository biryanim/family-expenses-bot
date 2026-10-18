package repository_test

import (
	"context"
	"errors"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/mock"
)

const listUsersSQL = "SELECT telegram_id, username, first_name, last_name, created_at, updated_at FROM users ORDER BY telegram_id"

func (s *Suite) TestListUsers_Success() {
	ctx := context.Background()

	mockRows := mock.NewMockRows(s.ctrl)
	gomock.InOrder(
		mockRows.EXPECT().Next().Return(true),
		mockRows.EXPECT().
			Scan(mockAnyArgs(6)...).
			DoAndReturn(func(dest ...any) error {
				*dest[0].(*int64) = 50
				*dest[1].(*string) = ""
				*dest[2].(*string) = "Анна"
				*dest[3].(*string) = ""
				*dest[4].(*time.Time) = createdAt
				*dest[5].(**time.Time) = nil
				return nil
			}),
		mockRows.EXPECT().Next().Return(true),
		mockRows.EXPECT().
			Scan(mockAnyArgs(6)...).
			DoAndReturn(func(dest ...any) error {
				*dest[0].(*int64) = 100
				*dest[1].(*string) = "petr"
				*dest[2].(*string) = "Пётр"
				*dest[3].(*string) = "Иванов"
				*dest[4].(*time.Time) = createdAt
				*dest[5].(**time.Time) = &updatedAt
				return nil
			}),
		mockRows.EXPECT().Next().Return(false),
	)
	mockRows.EXPECT().Err().Return(nil)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), listUsersSQL).
		Return(mockRows, nil)

	result, err := s.repo.ListUsers(ctx)

	s.NoError(err)
	s.Equal([]entity.User{
		{TelegramID: 50, FirstName: "Анна", CreatedAt: createdAt},
		{TelegramID: 100, Username: "petr", FirstName: "Пётр", LastName: "Иванов", CreatedAt: createdAt, UpdatedAt: &updatedAt},
	}, result)
}

func (s *Suite) TestListUsers_Empty() {
	ctx := context.Background()

	mockRows := mock.NewMockRows(s.ctrl)
	mockRows.EXPECT().Next().Return(false)
	mockRows.EXPECT().Err().Return(nil)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), listUsersSQL).
		Return(mockRows, nil)

	result, err := s.repo.ListUsers(ctx)

	s.NoError(err)
	s.NotNil(result)
	s.Empty(result)
}

func (s *Suite) TestListUsers_QueryError() {
	ctx := context.Background()
	queryErr := errors.New("query error")

	s.pool.EXPECT().
		Query(gomock.Any(), gomock.Any()).
		Return(nil, queryErr)

	result, err := s.repo.ListUsers(ctx)

	s.Error(err)
	s.Nil(result)
	s.ErrorIs(err, queryErr)
	s.Contains(err.Error(), "r.pool.Query")
}

func (s *Suite) TestListUsers_ScanError() {
	ctx := context.Background()
	scanErr := errors.New("scan error")

	mockRows := mock.NewMockRows(s.ctrl)
	mockRows.EXPECT().Next().Return(true)
	mockRows.EXPECT().
		Scan(mockAnyArgs(6)...).
		Return(scanErr)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), gomock.Any()).
		Return(mockRows, nil)

	result, err := s.repo.ListUsers(ctx)

	s.Error(err)
	s.Nil(result)
	s.ErrorIs(err, scanErr)
	s.Contains(err.Error(), "rows.Scan")
}

func (s *Suite) TestListUsers_RowsError() {
	ctx := context.Background()
	rowsErr := errors.New("rows error")

	mockRows := mock.NewMockRows(s.ctrl)
	mockRows.EXPECT().Next().Return(false)
	mockRows.EXPECT().Err().Return(rowsErr)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), gomock.Any()).
		Return(mockRows, nil)

	result, err := s.repo.ListUsers(ctx)

	s.Error(err)
	s.Nil(result)
	s.ErrorIs(err, rowsErr)
	s.Contains(err.Error(), "rows.Err")
}
