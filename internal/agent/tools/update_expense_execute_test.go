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

func (s *Suite) TestUpdateExpenseExecute_AmountOnly() {
	ctx := context.Background()
	expected := testExpense()
	expected.Amount = 45000

	s.uc.EXPECT().
		GetExpenseForUser(gomock.Any(), userA, int64(7)).
		Return(testExpense(), nil)
	s.uc.EXPECT().
		UpdateExpenseForUser(gomock.Any(), userA, expected).
		Return(expected, nil)

	result, err := s.updateExpense.Execute(ctx, userA, `{"id":7,"amount":450}`)

	s.NoError(err)
	s.Equal(tools.Expense{ID: 7, Amount: "450", Category: entity.CategoryFood, Date: "2026-10-17", Comment: "кофе"}, result)
}

func (s *Suite) TestUpdateExpenseExecute_AllFields() {
	ctx := context.Background()
	expected := testExpense()
	expected.Amount = 120000
	expected.Category = entity.CategoryTransport
	expected.Comment = "такси"
	expected.SpentAt = time.Date(2026, 10, 16, 0, 0, 0, 0, time.Local)

	s.uc.EXPECT().
		GetExpenseForUser(gomock.Any(), userA, int64(7)).
		Return(testExpense(), nil)
	s.uc.EXPECT().
		UpdateExpenseForUser(gomock.Any(), userA, expected).
		Return(expected, nil)

	result, err := s.updateExpense.Execute(ctx, userA, `{"id":7,"amount":1200,"category":"transport","comment":"такси","date":"2026-10-16"}`)

	s.NoError(err)
	s.Equal(tools.Expense{ID: 7, Amount: "1200", Category: entity.CategoryTransport, Date: "2026-10-16", Comment: "такси"}, result)
}

func (s *Suite) TestUpdateExpenseExecute_CategoryFallback() {
	ctx := context.Background()
	expected := testExpense()
	expected.Category = entity.CategoryOther

	s.uc.EXPECT().
		GetExpenseForUser(gomock.Any(), userA, int64(7)).
		Return(testExpense(), nil)
	s.uc.EXPECT().
		UpdateExpenseForUser(gomock.Any(), userA, expected).
		Return(expected, nil)

	_, err := s.updateExpense.Execute(ctx, userA, `{"id":7,"category":"travel"}`)

	s.NoError(err)
}

func (s *Suite) TestUpdateExpenseExecute_OtherUsersExpense() {
	ctx := context.Background()
	idOfUserB := int64(8)

	s.uc.EXPECT().
		GetExpenseForUser(gomock.Any(), userA, idOfUserB).
		Return(entity.Expense{}, entity.ErrNotFound)

	result, err := s.updateExpense.Execute(ctx, userA, `{"id":8,"amount":450}`)

	s.NoError(err)
	s.Equal(agent.ToolError{Error: "expense not found"}, result)
}

func (s *Suite) TestUpdateExpenseExecute_InvalidAmount() {
	ctx := context.Background()

	result, err := s.updateExpense.Execute(ctx, userA, `{"id":7,"amount":0}`)

	s.NoError(err)
	s.IsType(agent.ToolError{}, result)
}

func (s *Suite) TestUpdateExpenseExecute_InvalidDate() {
	ctx := context.Background()

	result, err := s.updateExpense.Execute(ctx, userA, `{"id":7,"date":"вчера"}`)

	s.NoError(err)
	s.Equal(agent.ToolError{Error: "invalid date: expected YYYY-MM-DD"}, result)
}

func (s *Suite) TestUpdateExpenseExecute_UserIDInArguments() {
	ctx := context.Background()

	result, err := s.updateExpense.Execute(ctx, userA, `{"id":8,"amount":450,"user_id":200}`)

	s.NoError(err)
	s.IsType(agent.ToolError{}, result)
	s.Contains(result.(agent.ToolError).Error, "invalid arguments")
}

func (s *Suite) TestUpdateExpenseExecute_GetError() {
	ctx := context.Background()
	ucErr := errors.New("connection refused")

	s.uc.EXPECT().
		GetExpenseForUser(gomock.Any(), userA, int64(7)).
		Return(entity.Expense{}, ucErr)

	result, err := s.updateExpense.Execute(ctx, userA, `{"id":7,"amount":450}`)

	s.Nil(result)
	s.ErrorIs(err, ucErr)
	s.Contains(err.Error(), "t.usecase.GetExpenseForUser")
}

func (s *Suite) TestUpdateExpenseExecute_UpdateError() {
	ctx := context.Background()
	ucErr := errors.New("connection refused")

	s.uc.EXPECT().
		GetExpenseForUser(gomock.Any(), userA, int64(7)).
		Return(testExpense(), nil)
	s.uc.EXPECT().
		UpdateExpenseForUser(gomock.Any(), userA, gomock.Any()).
		Return(entity.Expense{}, ucErr)

	result, err := s.updateExpense.Execute(ctx, userA, `{"id":7,"amount":450}`)

	s.Nil(result)
	s.ErrorIs(err, ucErr)
	s.Contains(err.Error(), "t.usecase.UpdateExpenseForUser")
}
