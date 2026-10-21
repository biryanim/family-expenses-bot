package repository_test

import (
	"context"
	"errors"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/mock"
)

func (s *Suite) TestListExpenses_Success() {
	ctx := context.Background()
	first := testExpense()
	second := testExpense()
	second.ID = 6
	second.Category = entity.CategoryTransport
	second.UpdatedAt = &updatedAt

	mockRows := mock.NewMockRows(s.ctrl)
	gomock.InOrder(
		mockRows.EXPECT().Next().Return(true),
		mockRows.EXPECT().
			Scan(pageScanArgs()...).
			DoAndReturn(func(dest ...any) error {
				populatePageScanDest(first, 45, 900000, dest...)
				return nil
			}),
		mockRows.EXPECT().Next().Return(true),
		mockRows.EXPECT().
			Scan(pageScanArgs()...).
			DoAndReturn(func(dest ...any) error {
				populatePageScanDest(second, 45, 900000, dest...)
				return nil
			}),
		mockRows.EXPECT().Next().Return(false),
	)
	mockRows.EXPECT().Err().Return(nil)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), selectPage+""+pageTail(0), []any{30, 0}...).
		Return(mockRows, nil)

	result, err := s.repo.ListExpenses(ctx, entity.ExpenseFilter{Limit: 30})

	s.NoError(err)
	s.Equal([]entity.Expense{first, second}, result.Items)
	s.Equal(int64(45), result.Count)
	s.Equal(int64(900000), result.Total)
}

func (s *Suite) TestListExpenses_Offset() {
	ctx := context.Background()

	mockRows := mock.NewMockRows(s.ctrl)
	mockRows.EXPECT().Next().Return(false)
	mockRows.EXPECT().Err().Return(nil)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), selectPage+""+pageTail(0), []any{30, 30}...).
		Return(mockRows, nil)
	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().
		Scan(gomock.Any(), gomock.Any()).
		DoAndReturn(func(dest ...any) error {
			*dest[0].(*int64) = 10
			*dest[1].(*int64) = 350000
			return nil
		})
	s.pool.EXPECT().
		QueryRow(gomock.Any(), selectTotals+"").
		Return(mockRow)

	result, err := s.repo.ListExpenses(ctx, entity.ExpenseFilter{Limit: 30, Offset: 30})

	s.NoError(err)
	s.NotNil(result.Items)
	s.Empty(result.Items)
	s.Equal(int64(10), result.Count)
	s.Equal(int64(350000), result.Total)
}

func (s *Suite) TestListExpenses_OffsetTotalsError() {
	ctx := context.Background()
	totalsErr := errors.New("totals error")

	mockRows := mock.NewMockRows(s.ctrl)
	mockRows.EXPECT().Next().Return(false)
	mockRows.EXPECT().Err().Return(nil)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), selectPage+""+pageTail(0), []any{30, 30}...).
		Return(mockRows, nil)
	mockRow := mock.NewMockRow(s.ctrl)
	mockRow.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(totalsErr)
	s.pool.EXPECT().
		QueryRow(gomock.Any(), selectTotals+"").
		Return(mockRow)

	result, err := s.repo.ListExpenses(ctx, entity.ExpenseFilter{Limit: 30, Offset: 30})

	s.ErrorIs(err, totalsErr)
	s.Contains(err.Error(), "r.pool.QueryRow.Scan")
	s.Zero(result)
}

func (s *Suite) TestListExpenses_Empty() {
	ctx := context.Background()

	mockRows := mock.NewMockRows(s.ctrl)
	mockRows.EXPECT().Next().Return(false)
	mockRows.EXPECT().Err().Return(nil)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(mockRows, nil)

	result, err := s.repo.ListExpenses(ctx, entity.ExpenseFilter{Limit: 30})

	s.NoError(err)
	s.NotNil(result.Items)
	s.Empty(result.Items)
	s.Zero(result.Count)
	s.Zero(result.Total)
}

func (s *Suite) TestListExpenses_QueryError() {
	ctx := context.Background()
	queryErr := errors.New("query error")

	s.pool.EXPECT().
		Query(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, queryErr)

	result, err := s.repo.ListExpenses(ctx, entity.ExpenseFilter{Limit: 30})

	s.Error(err)
	s.Zero(result)
	s.ErrorIs(err, queryErr)
	s.Contains(err.Error(), "r.pool.Query")
}

func (s *Suite) TestListExpenses_ScanError() {
	ctx := context.Background()
	scanErr := errors.New("scan error")

	mockRows := mock.NewMockRows(s.ctrl)
	mockRows.EXPECT().Next().Return(true)
	mockRows.EXPECT().
		Scan(pageScanArgs()...).
		Return(scanErr)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(mockRows, nil)

	result, err := s.repo.ListExpenses(ctx, entity.ExpenseFilter{Limit: 30})

	s.Error(err)
	s.Zero(result)
	s.ErrorIs(err, scanErr)
	s.Contains(err.Error(), "rows.Scan")
}

func (s *Suite) TestListExpenses_RowsError() {
	ctx := context.Background()
	rowsErr := errors.New("rows error")

	mockRows := mock.NewMockRows(s.ctrl)
	mockRows.EXPECT().Next().Return(false)
	mockRows.EXPECT().Err().Return(rowsErr)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(mockRows, nil)

	result, err := s.repo.ListExpenses(ctx, entity.ExpenseFilter{Limit: 30})

	s.Error(err)
	s.Zero(result)
	s.ErrorIs(err, rowsErr)
	s.Contains(err.Error(), "rows.Err")
}

func (s *Suite) TestListExpenses_Filters() {
	user := userA

	tests := []struct {
		name      string
		filter    entity.ExpenseFilter
		wantWhere string
		wantArgs  []any
	}{
		{name: "none", filter: entity.ExpenseFilter{}, wantWhere: ""},
		{name: "user", filter: entity.ExpenseFilter{UserID: &user}, wantWhere: " WHERE user_id = $1", wantArgs: []any{userA}},
		{name: "from", filter: entity.ExpenseFilter{From: from}, wantWhere: " WHERE spent_at >= $1", wantArgs: []any{from}},
		{name: "to", filter: entity.ExpenseFilter{To: to}, wantWhere: " WHERE spent_at < $1", wantArgs: []any{to}},
		{name: "category", filter: entity.ExpenseFilter{Category: entity.CategoryFood}, wantWhere: " WHERE category = $1", wantArgs: []any{"food"}},
		{name: "user from", filter: entity.ExpenseFilter{UserID: &user, From: from}, wantWhere: " WHERE user_id = $1 AND spent_at >= $2", wantArgs: []any{userA, from}},
		{name: "user to", filter: entity.ExpenseFilter{UserID: &user, To: to}, wantWhere: " WHERE user_id = $1 AND spent_at < $2", wantArgs: []any{userA, to}},
		{name: "user category", filter: entity.ExpenseFilter{UserID: &user, Category: entity.CategoryFood}, wantWhere: " WHERE user_id = $1 AND category = $2", wantArgs: []any{userA, "food"}},
		{name: "from to", filter: entity.ExpenseFilter{From: from, To: to}, wantWhere: " WHERE spent_at >= $1 AND spent_at < $2", wantArgs: []any{from, to}},
		{name: "from category", filter: entity.ExpenseFilter{From: from, Category: entity.CategoryFood}, wantWhere: " WHERE spent_at >= $1 AND category = $2", wantArgs: []any{from, "food"}},
		{name: "to category", filter: entity.ExpenseFilter{To: to, Category: entity.CategoryFood}, wantWhere: " WHERE spent_at < $1 AND category = $2", wantArgs: []any{to, "food"}},
		{name: "user from to", filter: entity.ExpenseFilter{UserID: &user, From: from, To: to}, wantWhere: " WHERE user_id = $1 AND spent_at >= $2 AND spent_at < $3", wantArgs: []any{userA, from, to}},
		{name: "user from category", filter: entity.ExpenseFilter{UserID: &user, From: from, Category: entity.CategoryFood}, wantWhere: " WHERE user_id = $1 AND spent_at >= $2 AND category = $3", wantArgs: []any{userA, from, "food"}},
		{name: "user to category", filter: entity.ExpenseFilter{UserID: &user, To: to, Category: entity.CategoryFood}, wantWhere: " WHERE user_id = $1 AND spent_at < $2 AND category = $3", wantArgs: []any{userA, to, "food"}},
		{name: "from to category", filter: entity.ExpenseFilter{From: from, To: to, Category: entity.CategoryFood}, wantWhere: " WHERE spent_at >= $1 AND spent_at < $2 AND category = $3", wantArgs: []any{from, to, "food"}},
		{name: "user from to category", filter: entity.ExpenseFilter{UserID: &user, From: from, To: to, Category: entity.CategoryFood}, wantWhere: " WHERE user_id = $1 AND spent_at >= $2 AND spent_at < $3 AND category = $4", wantArgs: []any{userA, from, to, "food"}},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			mockRows := mock.NewMockRows(s.ctrl)
			mockRows.EXPECT().Next().Return(false)
			mockRows.EXPECT().Err().Return(nil)
			mockRows.EXPECT().Close()

			s.pool.EXPECT().
				Query(gomock.Any(), selectPage+tt.wantWhere+pageTail(len(tt.wantArgs)), append(tt.wantArgs, 30, 0)...).
				Return(mockRows, nil)

			tt.filter.Limit = 30
			result, err := s.repo.ListExpenses(ctx, tt.filter)

			s.NoError(err)
			s.Empty(result.Items)
		})
	}
}
