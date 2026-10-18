package repository_test

import (
	"context"
	"errors"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/mock"
)

const listUserIDsSQL = "SELECT telegram_id FROM users ORDER BY telegram_id"

func (s *Suite) TestListUserIDs_Success() {
	ctx := context.Background()

	mockRows := mock.NewMockRows(s.ctrl)
	gomock.InOrder(
		mockRows.EXPECT().Next().Return(true),
		mockRows.EXPECT().
			Scan(gomock.Any()).
			DoAndReturn(func(dest ...any) error {
				*dest[0].(*int64) = 50
				return nil
			}),
		mockRows.EXPECT().Next().Return(true),
		mockRows.EXPECT().
			Scan(gomock.Any()).
			DoAndReturn(func(dest ...any) error {
				*dest[0].(*int64) = 100
				return nil
			}),
		mockRows.EXPECT().Next().Return(false),
	)
	mockRows.EXPECT().Err().Return(nil)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), listUserIDsSQL).
		Return(mockRows, nil)

	result, err := s.repo.ListUserIDs(ctx)

	s.NoError(err)
	s.Equal([]int64{50, 100}, result)
}

func (s *Suite) TestListUserIDs_Empty() {
	ctx := context.Background()

	mockRows := mock.NewMockRows(s.ctrl)
	mockRows.EXPECT().Next().Return(false)
	mockRows.EXPECT().Err().Return(nil)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), listUserIDsSQL).
		Return(mockRows, nil)

	result, err := s.repo.ListUserIDs(ctx)

	s.NoError(err)
	s.NotNil(result)
	s.Empty(result)
}

func (s *Suite) TestListUserIDs_QueryError() {
	ctx := context.Background()
	queryErr := errors.New("query error")

	s.pool.EXPECT().
		Query(gomock.Any(), gomock.Any()).
		Return(nil, queryErr)

	result, err := s.repo.ListUserIDs(ctx)

	s.Error(err)
	s.Nil(result)
	s.ErrorIs(err, queryErr)
	s.Contains(err.Error(), "r.pool.Query")
}

func (s *Suite) TestListUserIDs_ScanError() {
	ctx := context.Background()
	scanErr := errors.New("scan error")

	mockRows := mock.NewMockRows(s.ctrl)
	mockRows.EXPECT().Next().Return(true)
	mockRows.EXPECT().
		Scan(gomock.Any()).
		Return(scanErr)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), gomock.Any()).
		Return(mockRows, nil)

	result, err := s.repo.ListUserIDs(ctx)

	s.Error(err)
	s.Nil(result)
	s.ErrorIs(err, scanErr)
	s.Contains(err.Error(), "rows.Scan")
}

func (s *Suite) TestListUserIDs_RowsError() {
	ctx := context.Background()
	rowsErr := errors.New("rows error")

	mockRows := mock.NewMockRows(s.ctrl)
	mockRows.EXPECT().Next().Return(false)
	mockRows.EXPECT().Err().Return(rowsErr)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), gomock.Any()).
		Return(mockRows, nil)

	result, err := s.repo.ListUserIDs(ctx)

	s.Error(err)
	s.Nil(result)
	s.ErrorIs(err, rowsErr)
	s.Contains(err.Error(), "rows.Err")
}
