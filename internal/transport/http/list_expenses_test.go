package http_test

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/internal/transport/http/openapi"
)

func (s *Suite) TestListExpenses_Success() {
	updated := testExpense()
	updated.ID = 8
	updated.UpdatedAt = &updatedAt

	s.usecase.EXPECT().
		ListExpenses(gomock.Any(), entity.ExpenseFilter{}).
		Return([]entity.Expense{updated, testExpense()}, nil)

	rec := s.serve(http.MethodGet, "/api/expenses", "")

	s.Equal(http.StatusOK, rec.Code)
	var body []openapi.Expense
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &body))
	s.Equal([]openapi.Expense{
		{Id: 8, UserId: userA, Amount: 35000, Category: openapi.Food, SpentAt: spentAt, Comment: "кофе", CreatedAt: createdAt, UpdatedAt: &updatedAt},
		{Id: 7, UserId: userA, Amount: 35000, Category: openapi.Food, SpentAt: spentAt, Comment: "кофе", CreatedAt: createdAt},
	}, body)
}

func (s *Suite) TestListExpenses_Empty() {
	s.usecase.EXPECT().
		ListExpenses(gomock.Any(), gomock.Any()).
		Return(nil, nil)

	rec := s.serve(http.MethodGet, "/api/expenses", "")

	s.Equal(http.StatusOK, rec.Code)
	s.JSONEq(`[]`, rec.Body.String())
}

func (s *Suite) TestListExpenses_Filters() {
	user := userA
	tests := []struct {
		name  string
		query string
		want  entity.ExpenseFilter
	}{
		{name: "none", query: "", want: entity.ExpenseFilter{}},
		{name: "user", query: "?user_id=100", want: entity.ExpenseFilter{UserID: &user}},
		{name: "from", query: "?from=2026-10-01T00:00:00Z", want: entity.ExpenseFilter{From: from}},
		{name: "to", query: "?to=2026-10-08T00:00:00Z", want: entity.ExpenseFilter{To: to}},
		{name: "category", query: "?category=food", want: entity.ExpenseFilter{Category: entity.CategoryFood}},
		{name: "user and period", query: "?user_id=100&from=2026-10-01T00:00:00Z&to=2026-10-08T00:00:00Z", want: entity.ExpenseFilter{UserID: &user, From: from, To: to}},
		{name: "user and category", query: "?user_id=100&category=transport", want: entity.ExpenseFilter{UserID: &user, Category: entity.CategoryTransport}},
		{name: "period and category", query: "?from=2026-10-01T00:00:00Z&to=2026-10-08T00:00:00Z&category=health", want: entity.ExpenseFilter{From: from, To: to, Category: entity.CategoryHealth}},
		{name: "all", query: "?user_id=100&from=2026-10-01T00:00:00Z&to=2026-10-08T00:00:00Z&category=food", want: entity.ExpenseFilter{UserID: &user, From: from, To: to, Category: entity.CategoryFood}},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.usecase.EXPECT().
				ListExpenses(gomock.Any(), tt.want).
				Return([]entity.Expense{}, nil)

			rec := s.serve(http.MethodGet, "/api/expenses"+tt.query, "")

			s.Equal(http.StatusOK, rec.Code)
		})
	}
}

func (s *Suite) TestListExpenses_InvalidUserID() {
	rec := s.serve(http.MethodGet, "/api/expenses?user_id=abc", "")

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Equal("application/json", rec.Header().Get("Content-Type"))
	s.Contains(rec.Body.String(), `"code":"invalid_request"`)
}

func (s *Suite) TestListExpenses_InvalidFrom() {
	rec := s.serve(http.MethodGet, "/api/expenses?from=yesterday", "")

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Contains(rec.Body.String(), `"code":"invalid_request"`)
}

func (s *Suite) TestListExpenses_InvalidCategory() {
	s.usecase.EXPECT().
		ListExpenses(gomock.Any(), entity.ExpenseFilter{Category: "travel"}).
		Return(nil, entity.ErrInvalidCategory)

	rec := s.serve(http.MethodGet, "/api/expenses?category=travel", "")

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Contains(rec.Body.String(), `"code":"invalid_category"`)
}

func (s *Suite) TestListExpenses_InvalidPeriod() {
	s.usecase.EXPECT().
		ListExpenses(gomock.Any(), entity.ExpenseFilter{From: to, To: from}).
		Return(nil, entity.ErrInvalidDate)

	rec := s.serve(http.MethodGet, "/api/expenses?from=2026-10-08T00:00:00Z&to=2026-10-01T00:00:00Z", "")

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Contains(rec.Body.String(), `"code":"invalid_date"`)
}

func (s *Suite) TestListExpenses_InternalError() {
	s.usecase.EXPECT().
		ListExpenses(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("u.repo.ListExpenses: connection refused"))

	rec := s.serve(http.MethodGet, "/api/expenses", "")

	s.Equal(http.StatusInternalServerError, rec.Code)
	s.Contains(rec.Body.String(), `"code":"internal"`)
	s.NotContains(rec.Body.String(), "connection refused")
}
