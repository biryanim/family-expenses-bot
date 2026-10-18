package tools_test

import (
	"context"
	"errors"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/agent"
	"github.com/biryanim/family-expenses-bot/internal/agent/tools"
	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestDeleteExpenseExecute_Success() {
	ctx := context.Background()

	s.uc.EXPECT().
		GetExpenseForUser(gomock.Any(), userA, int64(7)).
		Return(testExpense(), nil)
	s.uc.EXPECT().
		DeleteExpenseForUser(gomock.Any(), userA, int64(7)).
		Return(nil)

	result, err := s.deleteExpense.Execute(ctx, userA, `{"id":7}`)

	s.NoError(err)
	s.Equal(tools.Expense{ID: 7, Amount: "350", Category: entity.CategoryFood, Date: "2026-10-17", Comment: "кофе"}, result)
}

func (s *Suite) TestDeleteExpenseExecute_OtherUsersExpense() {
	ctx := context.Background()
	idOfUserB := int64(8)

	s.uc.EXPECT().
		GetExpenseForUser(gomock.Any(), userA, idOfUserB).
		Return(entity.Expense{}, entity.ErrNotFound)

	result, err := s.deleteExpense.Execute(ctx, userA, `{"id":8}`)

	s.NoError(err)
	s.Equal(agent.ToolError{Error: "expense not found"}, result)
}

func (s *Suite) TestDeleteExpenseExecute_UserIDInArguments() {
	ctx := context.Background()

	result, err := s.deleteExpense.Execute(ctx, userA, `{"id":8,"user_id":200}`)

	s.NoError(err)
	s.IsType(agent.ToolError{}, result)
	s.Contains(result.(agent.ToolError).Error, "invalid arguments")
}

func (s *Suite) TestDeleteExpenseExecute_DeleteError() {
	ctx := context.Background()
	ucErr := errors.New("connection refused")

	s.uc.EXPECT().
		GetExpenseForUser(gomock.Any(), userA, int64(7)).
		Return(testExpense(), nil)
	s.uc.EXPECT().
		DeleteExpenseForUser(gomock.Any(), userA, int64(7)).
		Return(ucErr)

	result, err := s.deleteExpense.Execute(ctx, userA, `{"id":7}`)

	s.Nil(result)
	s.ErrorIs(err, ucErr)
	s.Contains(err.Error(), "t.usecase.DeleteExpenseForUser")
}
