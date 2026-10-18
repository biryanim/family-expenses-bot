package http_test

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/internal/transport/http/openapi"
)

func (s *Suite) TestGetExpense_Success() {
	s.usecase.EXPECT().
		GetExpense(gomock.Any(), int64(7)).
		Return(testExpense(), nil)

	rec := s.serve(http.MethodGet, "/api/expenses/7", "")

	s.Equal(http.StatusOK, rec.Code)
	var body openapi.Expense
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &body))
	s.Equal(openapi.Expense{Id: 7, UserId: userA, Amount: 35000, Category: openapi.Food, SpentAt: spentAt, Comment: "кофе", CreatedAt: createdAt}, body)
}

func (s *Suite) TestGetExpense_InvalidID() {
	rec := s.serve(http.MethodGet, "/api/expenses/abc", "")

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Contains(rec.Body.String(), `"code":"invalid_request"`)
}

func (s *Suite) TestGetExpense_NotFound() {
	s.usecase.EXPECT().
		GetExpense(gomock.Any(), int64(999)).
		Return(entity.Expense{}, entity.ErrNotFound)

	rec := s.serve(http.MethodGet, "/api/expenses/999", "")

	s.Equal(http.StatusNotFound, rec.Code)
	s.Contains(rec.Body.String(), `"code":"not_found"`)
}

func (s *Suite) TestGetExpense_InternalError() {
	s.usecase.EXPECT().
		GetExpense(gomock.Any(), int64(7)).
		Return(entity.Expense{}, errors.New("u.repo.GetExpense: connection refused"))

	rec := s.serve(http.MethodGet, "/api/expenses/7", "")

	s.Equal(http.StatusInternalServerError, rec.Code)
	s.Contains(rec.Body.String(), `"code":"internal"`)
	s.NotContains(rec.Body.String(), "connection refused")
}
