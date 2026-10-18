package repository_test

import (
	"context"
	"errors"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/mock"
)

func (s *Suite) TestListExpensesForUser_Success() {
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
		Query(gomock.Any(), selectExpenses+" WHERE user_id = $1"+orderExpenses, userA).
		Return(mockRows, nil)

	result, err := s.repo.ListExpensesForUser(ctx, userA, entity.ExpenseFilter{})

	s.NoError(err)
	s.Len(result, 2)
	s.Equal(first, result[0])
	s.Equal(second, result[1])
}

func (s *Suite) TestListExpensesForUser_Empty() {
	ctx := context.Background()

	mockRows := mock.NewMockRows(s.ctrl)
	mockRows.EXPECT().Next().Return(false)
	mockRows.EXPECT().Err().Return(nil)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), gomock.Any(), userA).
		Return(mockRows, nil)

	result, err := s.repo.ListExpensesForUser(ctx, userA, entity.ExpenseFilter{})

	s.NoError(err)
	s.NotNil(result)
	s.Empty(result)
}

func (s *Suite) TestListExpensesForUser_QueryError() {
	ctx := context.Background()
	queryErr := errors.New("query error")

	s.pool.EXPECT().
		Query(gomock.Any(), gomock.Any(), userA).
		Return(nil, queryErr)

	result, err := s.repo.ListExpensesForUser(ctx, userA, entity.ExpenseFilter{})

	s.Error(err)
	s.Nil(result)
	s.ErrorIs(err, queryErr)
	s.Contains(err.Error(), "r.pool.Query")
}

func (s *Suite) TestListExpensesForUser_ScanError() {
	ctx := context.Background()
	scanErr := errors.New("scan error")

	mockRows := mock.NewMockRows(s.ctrl)
	mockRows.EXPECT().Next().Return(true)
	mockRows.EXPECT().
		Scan(expenseScanArgs()...).
		Return(scanErr)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), gomock.Any(), userA).
		Return(mockRows, nil)

	result, err := s.repo.ListExpensesForUser(ctx, userA, entity.ExpenseFilter{})

	s.Error(err)
	s.Nil(result)
	s.ErrorIs(err, scanErr)
	s.Contains(err.Error(), "rows.Scan")
}

func (s *Suite) TestListExpensesForUser_RowsError() {
	ctx := context.Background()
	rowsErr := errors.New("rows error")

	mockRows := mock.NewMockRows(s.ctrl)
	mockRows.EXPECT().Next().Return(false)
	mockRows.EXPECT().Err().Return(rowsErr)
	mockRows.EXPECT().Close()

	s.pool.EXPECT().
		Query(gomock.Any(), gomock.Any(), userA).
		Return(mockRows, nil)

	result, err := s.repo.ListExpensesForUser(ctx, userA, entity.ExpenseFilter{})

	s.Error(err)
	s.Nil(result)
	s.ErrorIs(err, rowsErr)
	s.Contains(err.Error(), "rows.Err")
}

func (s *Suite) TestListExpensesForUser_Filters() {
	other := userB

	tests := []struct {
		name     string
		filter   entity.ExpenseFilter
		wantSQL  string
		wantArgs []any
	}{
		{name: "none", filter: entity.ExpenseFilter{}, wantSQL: selectExpenses + " WHERE user_id = $1" + orderExpenses, wantArgs: []any{userA}},
		{name: "from", filter: entity.ExpenseFilter{From: from}, wantSQL: selectExpenses + " WHERE user_id = $1 AND spent_at >= $2" + orderExpenses, wantArgs: []any{userA, from}},
		{name: "to", filter: entity.ExpenseFilter{To: to}, wantSQL: selectExpenses + " WHERE user_id = $1 AND spent_at < $2" + orderExpenses, wantArgs: []any{userA, to}},
		{name: "category", filter: entity.ExpenseFilter{Category: entity.CategoryFood}, wantSQL: selectExpenses + " WHERE user_id = $1 AND category = $2" + orderExpenses, wantArgs: []any{userA, "food"}},
		{name: "from to", filter: entity.ExpenseFilter{From: from, To: to}, wantSQL: selectExpenses + " WHERE user_id = $1 AND spent_at >= $2 AND spent_at < $3" + orderExpenses, wantArgs: []any{userA, from, to}},
		{name: "from category", filter: entity.ExpenseFilter{From: from, Category: entity.CategoryFood}, wantSQL: selectExpenses + " WHERE user_id = $1 AND spent_at >= $2 AND category = $3" + orderExpenses, wantArgs: []any{userA, from, "food"}},
		{name: "to category", filter: entity.ExpenseFilter{To: to, Category: entity.CategoryFood}, wantSQL: selectExpenses + " WHERE user_id = $1 AND spent_at < $2 AND category = $3" + orderExpenses, wantArgs: []any{userA, to, "food"}},
		{name: "from to category", filter: entity.ExpenseFilter{From: from, To: to, Category: entity.CategoryFood}, wantSQL: selectExpenses + " WHERE user_id = $1 AND spent_at >= $2 AND spent_at < $3 AND category = $4" + orderExpenses, wantArgs: []any{userA, from, to, "food"}},
		{name: "user B in filter none", filter: entity.ExpenseFilter{UserID: &other}, wantSQL: selectExpenses + " WHERE user_id = $1" + orderExpenses, wantArgs: []any{userA}},
		{name: "user B in filter from", filter: entity.ExpenseFilter{UserID: &other, From: from}, wantSQL: selectExpenses + " WHERE user_id = $1 AND spent_at >= $2" + orderExpenses, wantArgs: []any{userA, from}},
		{name: "user B in filter to", filter: entity.ExpenseFilter{UserID: &other, To: to}, wantSQL: selectExpenses + " WHERE user_id = $1 AND spent_at < $2" + orderExpenses, wantArgs: []any{userA, to}},
		{name: "user B in filter category", filter: entity.ExpenseFilter{UserID: &other, Category: entity.CategoryFood}, wantSQL: selectExpenses + " WHERE user_id = $1 AND category = $2" + orderExpenses, wantArgs: []any{userA, "food"}},
		{name: "user B in filter from to", filter: entity.ExpenseFilter{UserID: &other, From: from, To: to}, wantSQL: selectExpenses + " WHERE user_id = $1 AND spent_at >= $2 AND spent_at < $3" + orderExpenses, wantArgs: []any{userA, from, to}},
		{name: "user B in filter from category", filter: entity.ExpenseFilter{UserID: &other, From: from, Category: entity.CategoryFood}, wantSQL: selectExpenses + " WHERE user_id = $1 AND spent_at >= $2 AND category = $3" + orderExpenses, wantArgs: []any{userA, from, "food"}},
		{name: "user B in filter to category", filter: entity.ExpenseFilter{UserID: &other, To: to, Category: entity.CategoryFood}, wantSQL: selectExpenses + " WHERE user_id = $1 AND spent_at < $2 AND category = $3" + orderExpenses, wantArgs: []any{userA, to, "food"}},
		{name: "user B in filter from to category", filter: entity.ExpenseFilter{UserID: &other, From: from, To: to, Category: entity.CategoryFood}, wantSQL: selectExpenses + " WHERE user_id = $1 AND spent_at >= $2 AND spent_at < $3 AND category = $4" + orderExpenses, wantArgs: []any{userA, from, to, "food"}},
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

			result, err := s.repo.ListExpensesForUser(ctx, userA, tt.filter)

			s.NoError(err)
			s.Empty(result)
		})
	}
}
