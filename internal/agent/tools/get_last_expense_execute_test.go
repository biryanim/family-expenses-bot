package tools_test

import (
	"context"
	"errors"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/agent"
	"github.com/biryanim/family-expenses-bot/internal/agent/tools"
	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestGetLastExpenseExecute_Success() {
	ctx := context.Background()

	s.uc.EXPECT().
		GetLastExpenseForUser(gomock.Any(), userA).
		Return(testExpense(), nil)

	result, err := s.getLastExpense.Execute(ctx, userA, `{}`)

	s.NoError(err)
	s.Equal(tools.Expense{ID: 7, Amount: "350", Category: entity.CategoryFood, Date: "2026-10-17", Comment: "кофе"}, result)
}

func (s *Suite) TestGetLastExpenseExecute_NotFound() {
	ctx := context.Background()

	s.uc.EXPECT().
		GetLastExpenseForUser(gomock.Any(), userB).
		Return(entity.Expense{}, entity.ErrNotFound)

	result, err := s.getLastExpense.Execute(ctx, userB, `{}`)

	s.NoError(err)
	s.Equal(agent.ToolError{Error: "expense not found"}, result)
}

func (s *Suite) TestGetLastExpenseExecute_UserIDInArguments() {
	ctx := context.Background()

	result, err := s.getLastExpense.Execute(ctx, userA, `{"user_id":456}`)

	s.NoError(err)
	s.IsType(agent.ToolError{}, result)
}

func (s *Suite) TestGetLastExpenseExecute_UsecaseError() {
	ctx := context.Background()
	ucErr := errors.New("connection refused")

	s.uc.EXPECT().
		GetLastExpenseForUser(gomock.Any(), userA).
		Return(entity.Expense{}, ucErr)

	result, err := s.getLastExpense.Execute(ctx, userA, `{}`)

	s.Nil(result)
	s.ErrorIs(err, ucErr)
	s.Contains(err.Error(), "t.usecase.GetLastExpenseForUser")
}
