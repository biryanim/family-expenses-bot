package http_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/internal/transport/http/openapi"
)

func (s *Suite) TestUpdateExpense_Success() {
	updated := testExpense()
	updated.Amount = 150000
	updated.Category = entity.CategoryTransport
	updated.Comment = "такси"
	updated.UpdatedAt = &updatedAt

	s.usecase.EXPECT().
		UpdateExpense(gomock.Any(), entity.Expense{ID: 7, Amount: 150000, Category: entity.CategoryTransport, SpentAt: spentAt, Comment: "такси"}).
		Return(updated, nil)

	rec := s.serve(http.MethodPut, "/api/expenses/7", `{"amount":150000,"category":"transport","spent_at":"2026-10-17T09:30:00Z","comment":"такси"}`)

	s.Equal(http.StatusOK, rec.Code)
	var body openapi.Expense
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &body))
	s.Equal(openapi.Expense{Id: 7, UserId: userA, Amount: 150000, Category: openapi.Transport, SpentAt: spentAt, Comment: "такси", CreatedAt: createdAt, UpdatedAt: &updatedAt}, body)
}

func (s *Suite) TestUpdateExpense_EmptyComment() {
	updated := testExpense()
	updated.Comment = ""
	updated.UpdatedAt = &updatedAt

	s.usecase.EXPECT().
		UpdateExpense(gomock.Any(), entity.Expense{ID: 7, Amount: 35000, Category: entity.CategoryFood, SpentAt: spentAt}).
		Return(updated, nil)

	rec := s.serve(http.MethodPut, "/api/expenses/7", `{"amount":35000,"category":"food","spent_at":"2026-10-17T09:30:00Z"}`)

	s.Equal(http.StatusOK, rec.Code)
	s.Contains(rec.Body.String(), `"comment":""`)
}

func (s *Suite) TestUpdateExpense_InvalidID() {
	rec := s.serve(http.MethodPut, "/api/expenses/abc", `{"amount":35000,"category":"food","spent_at":"2026-10-17T09:30:00Z"}`)

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Contains(rec.Body.String(), `"code":"invalid_request"`)
}

func (s *Suite) TestUpdateExpense_InvalidJSON() {
	rec := s.serve(http.MethodPut, "/api/expenses/7", `not json`)

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Contains(rec.Body.String(), `"code":"invalid_request"`)
}

func (s *Suite) TestUpdateExpense_InvalidAmount() {
	s.usecase.EXPECT().
		UpdateExpense(gomock.Any(), gomock.Any()).
		Return(entity.Expense{}, entity.ErrInvalidAmount)

	rec := s.serve(http.MethodPut, "/api/expenses/7", `{"amount":-100,"category":"food","spent_at":"2026-10-17T09:30:00Z"}`)

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Contains(rec.Body.String(), `"code":"invalid_amount"`)
}

func (s *Suite) TestUpdateExpense_InvalidCategory() {
	s.usecase.EXPECT().
		UpdateExpense(gomock.Any(), entity.Expense{ID: 1, Amount: 35000, Category: "travel", SpentAt: spentAt}).
		Return(entity.Expense{}, entity.ErrInvalidCategory)

	rec := s.serve(http.MethodPut, "/api/expenses/1", `{"amount":35000,"category":"travel","spent_at":"2026-10-17T09:30:00Z"}`)

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Contains(rec.Body.String(), `"code":"invalid_category"`)
}

func (s *Suite) TestUpdateExpense_InvalidDate() {
	s.usecase.EXPECT().
		UpdateExpense(gomock.Any(), entity.Expense{ID: 1, Amount: 35000, Category: entity.CategoryFood, SpentAt: time.Time{}}).
		Return(entity.Expense{}, entity.ErrInvalidDate)

	rec := s.serve(http.MethodPut, "/api/expenses/1", `{"amount":35000,"category":"food"}`)

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Contains(rec.Body.String(), `"code":"invalid_date"`)
}

func (s *Suite) TestUpdateExpense_NotFound() {
	s.usecase.EXPECT().
		UpdateExpense(gomock.Any(), gomock.Any()).
		Return(entity.Expense{}, entity.ErrNotFound)

	rec := s.serve(http.MethodPut, "/api/expenses/999", `{"amount":35000,"category":"food","spent_at":"2026-10-17T09:30:00Z"}`)

	s.Equal(http.StatusNotFound, rec.Code)
	s.Contains(rec.Body.String(), `"code":"not_found"`)
}

func (s *Suite) TestUpdateExpense_InternalError() {
	s.usecase.EXPECT().
		UpdateExpense(gomock.Any(), gomock.Any()).
		Return(entity.Expense{}, errors.New("u.repo.UpdateExpense: connection refused"))

	rec := s.serve(http.MethodPut, "/api/expenses/7", `{"amount":35000,"category":"food","spent_at":"2026-10-17T09:30:00Z"}`)

	s.Equal(http.StatusInternalServerError, rec.Code)
	s.Contains(rec.Body.String(), `"code":"internal"`)
	s.NotContains(rec.Body.String(), "connection refused")
}
