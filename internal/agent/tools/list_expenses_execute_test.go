package tools_test

import (
	"context"
	"errors"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/agent"
	"github.com/biryanim/family-expenses-bot/internal/agent/tools"
	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestListExpensesExecute_Success() {
	ctx := context.Background()
	taxi := testExpense()
	taxi.ID = 8
	taxi.Amount = 120050
	taxi.Category = entity.CategoryTransport
	taxi.Comment = "такси"
	filter := entity.ExpenseFilter{
		From:     time.Date(2026, 10, 1, 0, 0, 0, 0, time.Local),
		To:       time.Date(2026, 10, 19, 0, 0, 0, 0, time.Local),
		Category: entity.CategoryFood,
	}

	s.uc.EXPECT().
		ListExpensesForUser(gomock.Any(), userA, filter).
		Return([]entity.Expense{taxi, testExpense()}, nil)

	result, err := s.listExpenses.Execute(ctx, userA, `{"from":"2026-10-01","to":"2026-10-18","category":"food"}`)

	s.NoError(err)
	s.Equal(tools.ExpenseList{
		Count: 2,
		Total: "1550.50",
		Expenses: []tools.Expense{
			{ID: 8, Amount: "1200.50", Category: entity.CategoryTransport, Date: "2026-10-17", Comment: "такси"},
			{ID: 7, Amount: "350", Category: entity.CategoryFood, Date: "2026-10-17", Comment: "кофе"},
		},
	}, result)
}

func (s *Suite) TestListExpensesExecute_NoFilter() {
	ctx := context.Background()

	s.uc.EXPECT().
		ListExpensesForUser(gomock.Any(), userA, entity.ExpenseFilter{}).
		Return([]entity.Expense{}, nil)

	_, err := s.listExpenses.Execute(ctx, userA, `{}`)

	s.NoError(err)
}

func (s *Suite) TestListExpensesExecute_Empty() {
	ctx := context.Background()

	s.uc.EXPECT().
		ListExpensesForUser(gomock.Any(), userA, gomock.Any()).
		Return([]entity.Expense{}, nil)

	result, err := s.listExpenses.Execute(ctx, userA, `{"category":"health"}`)

	s.NoError(err)
	s.Equal(tools.ExpenseList{Count: 0, Total: "0", Expenses: []tools.Expense{}}, result)
}

func (s *Suite) TestListExpensesExecute_Truncated() {
	ctx := context.Background()
	expenses := make([]entity.Expense, 45)
	for i := range expenses {
		expenses[i] = testExpense()
		expenses[i].ID = int64(i + 1)
		expenses[i].Amount = 10000
	}

	s.uc.EXPECT().
		ListExpensesForUser(gomock.Any(), userA, gomock.Any()).
		Return(expenses, nil)

	result, err := s.listExpenses.Execute(ctx, userA, `{}`)

	s.NoError(err)
	list := result.(tools.ExpenseList)
	s.Equal(45, list.Count)
	s.Equal("4500", list.Total)
	s.Len(list.Expenses, 30)
	s.Equal(int64(1), list.Expenses[0].ID)
	s.Equal(int64(30), list.Expenses[29].ID)
	s.True(list.Truncated)
}

func (s *Suite) TestListExpensesExecute_InvalidCategory() {
	ctx := context.Background()

	s.uc.EXPECT().
		ListExpensesForUser(gomock.Any(), userA, entity.ExpenseFilter{Category: "travel"}).
		Return(nil, entity.ErrInvalidCategory)

	result, err := s.listExpenses.Execute(ctx, userA, `{"category":"travel"}`)

	s.NoError(err)
	s.Equal(agent.ToolError{Error: "invalid category"}, result)
}

func (s *Suite) TestListExpensesExecute_InvalidPeriod() {
	ctx := context.Background()

	s.uc.EXPECT().
		ListExpensesForUser(gomock.Any(), userA, gomock.Any()).
		Return(nil, entity.ErrInvalidDate)

	result, err := s.listExpenses.Execute(ctx, userA, `{"from":"2026-10-18","to":"2026-10-01"}`)

	s.NoError(err)
	s.Equal(agent.ToolError{Error: "invalid period: from is after to"}, result)
}

func (s *Suite) TestListExpensesExecute_InvalidDate() {
	for name, arguments := range map[string]string{
		"from": `{"from":"01.10.2026"}`,
		"to":   `{"to":"2026-13-01"}`,
	} {
		s.Run(name, func() {
			ctx := context.Background()

			result, err := s.listExpenses.Execute(ctx, userA, arguments)

			s.NoError(err)
			s.IsType(agent.ToolError{}, result)
		})
	}
}

func (s *Suite) TestListExpensesExecute_UserIDInArguments() {
	ctx := context.Background()

	result, err := s.listExpenses.Execute(ctx, userA, `{"user_id":456}`)

	s.NoError(err)
	s.IsType(agent.ToolError{}, result)
	s.Contains(result.(agent.ToolError).Error, "invalid arguments")
}

func (s *Suite) TestListExpensesExecute_OnlyOwnExpenses() {
	ctx := context.Background()
	ownExpense := testExpense()
	ownExpense.UserID = userA

	s.uc.EXPECT().
		ListExpensesForUser(gomock.Any(), userA, gomock.Any()).
		Return([]entity.Expense{ownExpense}, nil)

	result, err := s.listExpenses.Execute(ctx, userA, `{"from":"2026-10-12","to":"2026-10-18"}`)

	s.NoError(err)
	list := result.(tools.ExpenseList)
	s.Equal(1, list.Count)
	s.Equal(int64(7), list.Expenses[0].ID)
}

func (s *Suite) TestListExpensesExecute_UsecaseError() {
	ctx := context.Background()
	ucErr := errors.New("connection refused")

	s.uc.EXPECT().
		ListExpensesForUser(gomock.Any(), userA, gomock.Any()).
		Return(nil, ucErr)

	result, err := s.listExpenses.Execute(ctx, userA, `{}`)

	s.Nil(result)
	s.ErrorIs(err, ucErr)
	s.Contains(err.Error(), "t.usecase.ListExpensesForUser")
}
