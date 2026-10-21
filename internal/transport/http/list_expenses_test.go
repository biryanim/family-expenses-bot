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
		ListExpenses(gomock.Any(), entity.ExpenseFilter{Limit: 100}).
		Return(entity.ExpensePage{Items: []entity.Expense{updated, testExpense()}, Count: 250, Total: 70000}, nil)

	rec := s.serve(http.MethodGet, "/api/expenses", "")

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("250", rec.Header().Get("X-Total-Count"))
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
		Return(entity.ExpensePage{}, nil)

	rec := s.serve(http.MethodGet, "/api/expenses", "")

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("0", rec.Header().Get("X-Total-Count"))
	s.JSONEq(`[]`, rec.Body.String())
}

func (s *Suite) TestListExpenses_Filters() {
	user := userA
	tests := []struct {
		name  string
		query string
		want  entity.ExpenseFilter
	}{
		{name: "none", query: "", want: entity.ExpenseFilter{Limit: 100}},
		{name: "user", query: "?user_id=100", want: entity.ExpenseFilter{UserID: &user, Limit: 100}},
		{name: "from", query: "?from=2026-10-01T00:00:00Z", want: entity.ExpenseFilter{From: from, Limit: 100}},
		{name: "to", query: "?to=2026-10-08T00:00:00Z", want: entity.ExpenseFilter{To: to, Limit: 100}},
		{name: "category", query: "?category=food", want: entity.ExpenseFilter{Category: entity.CategoryFood, Limit: 100}},
		{name: "user and period", query: "?user_id=100&from=2026-10-01T00:00:00Z&to=2026-10-08T00:00:00Z", want: entity.ExpenseFilter{UserID: &user, From: from, To: to, Limit: 100}},
		{name: "user and category", query: "?user_id=100&category=transport", want: entity.ExpenseFilter{UserID: &user, Category: entity.CategoryTransport, Limit: 100}},
		{name: "period and category", query: "?from=2026-10-01T00:00:00Z&to=2026-10-08T00:00:00Z&category=health", want: entity.ExpenseFilter{From: from, To: to, Category: entity.CategoryHealth, Limit: 100}},
		{name: "all", query: "?user_id=100&from=2026-10-01T00:00:00Z&to=2026-10-08T00:00:00Z&category=food", want: entity.ExpenseFilter{UserID: &user, From: from, To: to, Category: entity.CategoryFood, Limit: 100}},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.usecase.EXPECT().
				ListExpenses(gomock.Any(), tt.want).
				Return(entity.ExpensePage{}, nil)

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
		ListExpenses(gomock.Any(), entity.ExpenseFilter{Category: "travel", Limit: 100}).
		Return(entity.ExpensePage{}, entity.ErrInvalidCategory)

	rec := s.serve(http.MethodGet, "/api/expenses?category=travel", "")

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Contains(rec.Body.String(), `"code":"invalid_category"`)
}

func (s *Suite) TestListExpenses_InvalidPeriod() {
	s.usecase.EXPECT().
		ListExpenses(gomock.Any(), entity.ExpenseFilter{From: to, To: from, Limit: 100}).
		Return(entity.ExpensePage{}, entity.ErrInvalidDate)

	rec := s.serve(http.MethodGet, "/api/expenses?from=2026-10-08T00:00:00Z&to=2026-10-01T00:00:00Z", "")

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Contains(rec.Body.String(), `"code":"invalid_date"`)
}

func (s *Suite) TestListExpenses_InternalError() {
	s.usecase.EXPECT().
		ListExpenses(gomock.Any(), gomock.Any()).
		Return(entity.ExpensePage{}, errors.New("u.repo.ListExpenses: connection refused"))

	rec := s.serve(http.MethodGet, "/api/expenses", "")

	s.Equal(http.StatusInternalServerError, rec.Code)
	s.Contains(rec.Body.String(), `"code":"internal"`)
	s.NotContains(rec.Body.String(), "connection refused")
}

func (s *Suite) TestListExpenses_Pagination() {
	s.usecase.EXPECT().
		ListExpenses(gomock.Any(), entity.ExpenseFilter{Limit: 100, Offset: 100}).
		Return(entity.ExpensePage{Items: []entity.Expense{}, Count: 250}, nil)

	rec := s.serve(http.MethodGet, "/api/expenses?limit=100&offset=100", "")

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("250", rec.Header().Get("X-Total-Count"))
}

func (s *Suite) TestListExpenses_OffsetBeyondResults() {
	s.usecase.EXPECT().
		ListExpenses(gomock.Any(), entity.ExpenseFilter{Limit: 100, Offset: 100}).
		Return(entity.ExpensePage{Items: []entity.Expense{}, Count: 50}, nil)

	rec := s.serve(http.MethodGet, "/api/expenses?offset=100", "")

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("50", rec.Header().Get("X-Total-Count"))
	s.JSONEq(`[]`, rec.Body.String())
}

func (s *Suite) TestListExpenses_InvalidPagination() {
	for name, query := range map[string]string{
		"limit zero":      "?limit=0",
		"limit over max":  "?limit=501",
		"negative offset": "?offset=-1",
	} {
		s.Run(name, func() {
			s.usecase.EXPECT().
				ListExpenses(gomock.Any(), gomock.Any()).
				Return(entity.ExpensePage{}, entity.ErrInvalidPagination)

			rec := s.serve(http.MethodGet, "/api/expenses"+query, "")

			s.Equal(http.StatusBadRequest, rec.Code)
			s.Contains(rec.Body.String(), `"code":"invalid_pagination"`)
		})
	}
}

func (s *Suite) TestListExpenses_NonIntegerLimit() {
	rec := s.serve(http.MethodGet, "/api/expenses?limit=abc", "")

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Contains(rec.Body.String(), `"code":"invalid_request"`)
}
