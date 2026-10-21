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
	amount := int64(45000)
	updated := testExpense()
	updated.Amount = 45000

	s.uc.EXPECT().
		UpdateExpenseForUser(gomock.Any(), userA, int64(7), entity.ExpensePatch{Amount: &amount}).
		Return(updated, nil)

	result, err := s.updateExpense.Execute(ctx, userA, `{"id":7,"amount":450}`)

	s.NoError(err)
	s.Equal(tools.Expense{ID: 7, Amount: "450", Category: entity.CategoryFood, Date: "2026-10-17", Comment: "кофе"}, result)
}

func (s *Suite) TestUpdateExpenseExecute_AllFields() {
	ctx := context.Background()
	amount := int64(120000)
	category := entity.CategoryTransport
	comment := "такси"
	date := time.Date(2026, 10, 16, 0, 0, 0, 0, time.Local)
	updated := testExpense()
	updated.Amount = 120000
	updated.Category = entity.CategoryTransport
	updated.Comment = "такси"
	updated.SpentAt = date

	s.uc.EXPECT().
		UpdateExpenseForUser(gomock.Any(), userA, int64(7), entity.ExpensePatch{Amount: &amount, Category: &category, Comment: &comment, SpentAt: &date}).
		Return(updated, nil)

	result, err := s.updateExpense.Execute(ctx, userA, `{"id":7,"amount":1200,"category":"transport","comment":"такси","date":"2026-10-16"}`)

	s.NoError(err)
	s.Equal(tools.Expense{ID: 7, Amount: "1200", Category: entity.CategoryTransport, Date: "2026-10-16", Comment: "такси"}, result)
}

func (s *Suite) TestUpdateExpenseExecute_CategoryFallback() {
	ctx := context.Background()
	category := entity.CategoryOther
	updated := testExpense()
	updated.Category = entity.CategoryOther

	s.uc.EXPECT().
		UpdateExpenseForUser(gomock.Any(), userA, int64(7), entity.ExpensePatch{Category: &category}).
		Return(updated, nil)

	_, err := s.updateExpense.Execute(ctx, userA, `{"id":7,"category":"travel"}`)

	s.NoError(err)
}

func (s *Suite) TestUpdateExpenseExecute_CommentOnlyDoesNotTouchOtherFields() {
	ctx := context.Background()
	comment := "латте"

	s.uc.EXPECT().
		UpdateExpenseForUser(gomock.Any(), userA, int64(7), entity.ExpensePatch{Comment: &comment}).
		Return(testExpense(), nil)

	_, err := s.updateExpense.Execute(ctx, userA, `{"id":7,"comment":"  латте "}`)

	s.NoError(err)
}

func (s *Suite) TestUpdateExpenseExecute_EmptyPatch() {
	ctx := context.Background()

	result, err := s.updateExpense.Execute(ctx, userA, `{"id":7}`)

	s.NoError(err)
	s.IsType(agent.ToolError{}, result)
	s.Contains(result.(agent.ToolError).Error, "nothing to update")
}

func (s *Suite) TestUpdateExpenseExecute_OtherUsersExpense() {
	ctx := context.Background()
	amount := int64(45000)

	s.uc.EXPECT().
		UpdateExpenseForUser(gomock.Any(), userA, int64(8), entity.ExpensePatch{Amount: &amount}).
		Return(entity.Expense{}, entity.ErrNotFound)

	result, err := s.updateExpense.Execute(ctx, userA, `{"id":8,"amount":450}`)

	s.NoError(err)
	s.Equal(agent.ToolError{Error: "expense not found"}, result)
}

func (s *Suite) TestUpdateExpenseExecute_CommentTooLong() {
	ctx := context.Background()

	s.uc.EXPECT().
		UpdateExpenseForUser(gomock.Any(), userA, int64(7), gomock.Any()).
		Return(entity.Expense{}, entity.ErrInvalidComment)

	result, err := s.updateExpense.Execute(ctx, userA, `{"id":7,"comment":"длинный"}`)

	s.NoError(err)
	s.IsType(agent.ToolError{}, result)
	s.Contains(result.(agent.ToolError).Error, "comment too long")
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

func (s *Suite) TestUpdateExpenseExecute_UpdateError() {
	ctx := context.Background()
	ucErr := errors.New("connection refused")

	s.uc.EXPECT().
		UpdateExpenseForUser(gomock.Any(), userA, int64(7), gomock.Any()).
		Return(entity.Expense{}, ucErr)

	result, err := s.updateExpense.Execute(ctx, userA, `{"id":7,"amount":450}`)

	s.Nil(result)
	s.ErrorIs(err, ucErr)
	s.Contains(err.Error(), "t.usecase.UpdateExpenseForUser")
}
