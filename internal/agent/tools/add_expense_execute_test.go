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

func (s *Suite) TestAddExpenseExecute_Success() {
	ctx := context.Background()

	s.uc.EXPECT().
		CreateExpense(gomock.Any(), entity.Expense{UserID: userA, Amount: 35000, Category: entity.CategoryFood, Comment: "кофе"}).
		Return(testExpense(), nil)

	result, err := s.addExpense.Execute(ctx, userA, `{"amount":350,"category":"food","comment":"кофе"}`)

	s.NoError(err)
	s.Equal(tools.Expense{ID: 7, Amount: "350", Category: entity.CategoryFood, Date: "2026-10-17", Comment: "кофе"}, result)
}

func (s *Suite) TestAddExpenseExecute_Kopecks() {
	ctx := context.Background()
	var saved entity.Expense

	s.uc.EXPECT().
		CreateExpense(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, e entity.Expense) (entity.Expense, error) {
			saved = e
			return testExpense(), nil
		})

	_, err := s.addExpense.Execute(ctx, userA, `{"amount":199.9,"category":"food","comment":"обед"}`)

	s.NoError(err)
	s.Equal(int64(19990), saved.Amount)
}

func (s *Suite) TestAddExpenseExecute_AmountAsString() {
	ctx := context.Background()
	var saved entity.Expense

	s.uc.EXPECT().
		CreateExpense(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, e entity.Expense) (entity.Expense, error) {
			saved = e
			return testExpense(), nil
		})

	_, err := s.addExpense.Execute(ctx, userA, `{"amount":"1200","category":"transport","comment":"такси"}`)

	s.NoError(err)
	s.Equal(int64(120000), saved.Amount)
}

func (s *Suite) TestAddExpenseExecute_CategoryFallback() {
	for name, arguments := range map[string]string{
		"unknown": `{"amount":350,"category":"travel","comment":"билет"}`,
		"missing": `{"amount":350,"comment":"билет"}`,
	} {
		s.Run(name, func() {
			ctx := context.Background()
			var saved entity.Expense

			s.uc.EXPECT().
				CreateExpense(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, e entity.Expense) (entity.Expense, error) {
					saved = e
					return testExpense(), nil
				})

			_, err := s.addExpense.Execute(ctx, userA, arguments)

			s.NoError(err)
			s.Equal(entity.CategoryOther, saved.Category)
		})
	}
}

func (s *Suite) TestAddExpenseExecute_Yesterday() {
	ctx := context.Background()
	yesterday := time.Now().AddDate(0, 0, -1)
	var saved entity.Expense

	s.uc.EXPECT().
		CreateExpense(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, e entity.Expense) (entity.Expense, error) {
			saved = e
			return testExpense(), nil
		})

	_, err := s.addExpense.Execute(ctx, userA, `{"amount":1200,"category":"transport","comment":"такси","date":"`+yesterday.Format(time.DateOnly)+`"}`)

	s.NoError(err)
	s.Equal(time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.Local), saved.SpentAt)
	s.Equal(int64(120000), saved.Amount)
	s.Equal(entity.CategoryTransport, saved.Category)
}

func (s *Suite) TestAddExpenseExecute_Today() {
	ctx := context.Background()
	var saved entity.Expense

	s.uc.EXPECT().
		CreateExpense(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, e entity.Expense) (entity.Expense, error) {
			saved = e
			return testExpense(), nil
		})

	_, err := s.addExpense.Execute(ctx, userA, `{"amount":350,"category":"food","comment":"кофе","date":"`+time.Now().Format(time.DateOnly)+`"}`)

	s.NoError(err)
	s.True(saved.SpentAt.IsZero())
}

func (s *Suite) TestAddExpenseExecute_InvalidDate() {
	for _, date := range []string{"17.10.2026", "2026-02-30", "вчера"} {
		s.Run(date, func() {
			ctx := context.Background()

			result, err := s.addExpense.Execute(ctx, userA, `{"amount":350,"category":"food","comment":"кофе","date":"`+date+`"}`)

			s.NoError(err)
			s.Equal(agent.ToolError{Error: "invalid date: expected YYYY-MM-DD"}, result)
		})
	}
}

func (s *Suite) TestAddExpenseExecute_InvalidAmount() {
	for _, amount := range []string{"0", "-5", "1.005", `"350 руб"`} {
		s.Run(amount, func() {
			ctx := context.Background()

			result, err := s.addExpense.Execute(ctx, userA, `{"amount":`+amount+`,"category":"food","comment":"кофе"}`)

			s.NoError(err)
			s.IsType(agent.ToolError{}, result)
		})
	}
}

func (s *Suite) TestAddExpenseExecute_InvalidJSON() {
	ctx := context.Background()

	result, err := s.addExpense.Execute(ctx, userA, `{amount: 350`)

	s.NoError(err)
	s.IsType(agent.ToolError{}, result)
	s.Contains(result.(agent.ToolError).Error, "invalid arguments")
}

func (s *Suite) TestAddExpenseExecute_UserIDInArguments() {
	ctx := context.Background()

	result, err := s.addExpense.Execute(ctx, userA, `{"amount":350,"category":"food","comment":"кофе","user_id":456}`)

	s.NoError(err)
	s.IsType(agent.ToolError{}, result)
	s.Contains(result.(agent.ToolError).Error, "invalid arguments")
}

func (s *Suite) TestAddExpenseExecute_UsesCallerUserID() {
	ctx := context.Background()

	s.uc.EXPECT().
		CreateExpense(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, e entity.Expense) (entity.Expense, error) {
			s.Equal(userB, e.UserID)
			return testExpense(), nil
		})

	_, err := s.addExpense.Execute(ctx, userB, `{"amount":350,"category":"food","comment":"кофе"}`)

	s.NoError(err)
}

func (s *Suite) TestAddExpenseExecute_UsecaseError() {
	ctx := context.Background()
	ucErr := errors.New("connection refused")

	s.uc.EXPECT().
		CreateExpense(gomock.Any(), gomock.Any()).
		Return(entity.Expense{}, ucErr)

	result, err := s.addExpense.Execute(ctx, userA, `{"amount":350,"category":"food","comment":"кофе"}`)

	s.Nil(result)
	s.ErrorIs(err, ucErr)
	s.Contains(err.Error(), "t.usecase.CreateExpense")
}

func (s *Suite) TestAddExpenseExecute_CommentTooLong() {
	ctx := context.Background()

	s.uc.EXPECT().
		CreateExpense(gomock.Any(), gomock.Any()).
		Return(entity.Expense{}, entity.ErrInvalidComment)

	result, err := s.addExpense.Execute(ctx, userA, `{"amount":350,"category":"food","comment":"длинный"}`)

	s.NoError(err)
	s.IsType(agent.ToolError{}, result)
	s.Contains(result.(agent.ToolError).Error, "comment too long")
}
