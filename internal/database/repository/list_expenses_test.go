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
			Scan(expenseScanArgs()...).
			DoAndReturn(func(dest ...any) error {
				populateExpenseScanDest(first, dest...)
				return nil
			}),
		mockRows.EXPECT().Next().Return(true),
		mockRows.EXPECT().
			Scan(expenseScanArgs()...).
			DoAndReturn(func(dest ...any) error {
				populateExpenseScanDest(second, dest...)
				return nil
			}),
		mockRows.EXPECT().Next().Return(false),
	)
	mockRows.EXPECT().Err().Return(nil)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), selectExpenses+orderExpenses).
		Return(mockRows, nil)

	result, err := s.repo.ListExpenses(ctx, entity.ExpenseFilter{})

	s.NoError(err)
	s.Len(result, 2)
	s.Equal(first, result[0])
	s.Equal(second, result[1])
}

func (s *Suite) TestListExpenses_Empty() {
	ctx := context.Background()

	mockRows := mock.NewMockRows(s.ctrl)
	mockRows.EXPECT().Next().Return(false)
	mockRows.EXPECT().Err().Return(nil)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), gomock.Any()).
		Return(mockRows, nil)

	result, err := s.repo.ListExpenses(ctx, entity.ExpenseFilter{})

	s.NoError(err)
	s.NotNil(result)
	s.Empty(result)
}

func (s *Suite) TestListExpenses_QueryError() {
	ctx := context.Background()
	queryErr := errors.New("query error")

	s.pool.EXPECT().
		Query(gomock.Any(), gomock.Any()).
		Return(nil, queryErr)

	result, err := s.repo.ListExpenses(ctx, entity.ExpenseFilter{})

	s.Error(err)
	s.Nil(result)
	s.ErrorIs(err, queryErr)
	s.Contains(err.Error(), "r.pool.Query")
}

func (s *Suite) TestListExpenses_ScanError() {
	ctx := context.Background()
	scanErr := errors.New("scan error")

	mockRows := mock.NewMockRows(s.ctrl)
	mockRows.EXPECT().Next().Return(true)
	mockRows.EXPECT().
		Scan(expenseScanArgs()...).
		Return(scanErr)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), gomock.Any()).
		Return(mockRows, nil)

	result, err := s.repo.ListExpenses(ctx, entity.ExpenseFilter{})

	s.Error(err)
	s.Nil(result)
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
		Query(gomock.Any(), gomock.Any()).
		Return(mockRows, nil)

	result, err := s.repo.ListExpenses(ctx, entity.ExpenseFilter{})

	s.Error(err)
	s.Nil(result)
	s.ErrorIs(err, rowsErr)
	s.Contains(err.Error(), "rows.Err")
}

func (s *Suite) TestListExpenses_Filters() {
	user := userA

	tests := []struct {
		name     string
		filter   entity.ExpenseFilter
		wantSQL  string
		wantArgs []any
	}{
		{name: "none", filter: entity.ExpenseFilter{}, wantSQL: selectExpenses + orderExpenses},
		{name: "user", filter: entity.ExpenseFilter{UserID: &user}, wantSQL: selectExpenses + " WHERE user_id = $1" + orderExpenses, wantArgs: []any{userA}},
		{name: "from", filter: entity.ExpenseFilter{From: from}, wantSQL: selectExpenses + " WHERE spent_at >= $1" + orderExpenses, wantArgs: []any{from}},
		{name: "to", filter: entity.ExpenseFilter{To: to}, wantSQL: selectExpenses + " WHERE spent_at < $1" + orderExpenses, wantArgs: []any{to}},
		{name: "category", filter: entity.ExpenseFilter{Category: entity.CategoryFood}, wantSQL: selectExpenses + " WHERE category = $1" + orderExpenses, wantArgs: []any{"food"}},
		{name: "user from", filter: entity.ExpenseFilter{UserID: &user, From: from}, wantSQL: selectExpenses + " WHERE user_id = $1 AND spent_at >= $2" + orderExpenses, wantArgs: []any{userA, from}},
		{name: "user to", filter: entity.ExpenseFilter{UserID: &user, To: to}, wantSQL: selectExpenses + " WHERE user_id = $1 AND spent_at < $2" + orderExpenses, wantArgs: []any{userA, to}},
		{name: "user category", filter: entity.ExpenseFilter{UserID: &user, Category: entity.CategoryFood}, wantSQL: selectExpenses + " WHERE user_id = $1 AND category = $2" + orderExpenses, wantArgs: []any{userA, "food"}},
		{name: "from to", filter: entity.ExpenseFilter{From: from, To: to}, wantSQL: selectExpenses + " WHERE spent_at >= $1 AND spent_at < $2" + orderExpenses, wantArgs: []any{from, to}},
		{name: "from category", filter: entity.ExpenseFilter{From: from, Category: entity.CategoryFood}, wantSQL: selectExpenses + " WHERE spent_at >= $1 AND category = $2" + orderExpenses, wantArgs: []any{from, "food"}},
		{name: "to category", filter: entity.ExpenseFilter{To: to, Category: entity.CategoryFood}, wantSQL: selectExpenses + " WHERE spent_at < $1 AND category = $2" + orderExpenses, wantArgs: []any{to, "food"}},
		{name: "user from to", filter: entity.ExpenseFilter{UserID: &user, From: from, To: to}, wantSQL: selectExpenses + " WHERE user_id = $1 AND spent_at >= $2 AND spent_at < $3" + orderExpenses, wantArgs: []any{userA, from, to}},
		{name: "user from category", filter: entity.ExpenseFilter{UserID: &user, From: from, Category: entity.CategoryFood}, wantSQL: selectExpenses + " WHERE user_id = $1 AND spent_at >= $2 AND category = $3" + orderExpenses, wantArgs: []any{userA, from, "food"}},
		{name: "user to category", filter: entity.ExpenseFilter{UserID: &user, To: to, Category: entity.CategoryFood}, wantSQL: selectExpenses + " WHERE user_id = $1 AND spent_at < $2 AND category = $3" + orderExpenses, wantArgs: []any{userA, to, "food"}},
		{name: "from to category", filter: entity.ExpenseFilter{From: from, To: to, Category: entity.CategoryFood}, wantSQL: selectExpenses + " WHERE spent_at >= $1 AND spent_at < $2 AND category = $3" + orderExpenses, wantArgs: []any{from, to, "food"}},
		{name: "user from to category", filter: entity.ExpenseFilter{UserID: &user, From: from, To: to, Category: entity.CategoryFood}, wantSQL: selectExpenses + " WHERE user_id = $1 AND spent_at >= $2 AND spent_at < $3 AND category = $4" + orderExpenses, wantArgs: []any{userA, from, to, "food"}},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			mockRows := mock.NewMockRows(s.ctrl)
			mockRows.EXPECT().Next().Return(false)
			mockRows.EXPECT().Err().Return(nil)
			mockRows.EXPECT().Close()

			s.pool.EXPECT().
				Query(gomock.Any(), tt.wantSQL, tt.wantArgs...).
				Return(mockRows, nil)

			result, err := s.repo.ListExpenses(ctx, tt.filter)

			s.NoError(err)
			s.Empty(result)
		})
	}
}
