package http_test

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/internal/transport/http/openapi"
)

func (s *Suite) TestCreateExpense_Success() {
	s.usecase.EXPECT().
		CreateExpense(gomock.Any(), entity.Expense{UserID: userA, Amount: 35000, Category: entity.CategoryFood, SpentAt: spentAt, Comment: "кофе"}).
		Return(testExpense(), nil)

	rec := s.serve(http.MethodPost, "/api/expenses", `{"user_id":100,"amount":35000,"category":"food","spent_at":"2026-10-17T09:30:00Z","comment":"кофе"}`)

	s.Equal(http.StatusCreated, rec.Code)
	var body openapi.Expense
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &body))
	s.Equal(openapi.Expense{Id: 7, UserId: userA, Amount: 35000, Category: openapi.Food, SpentAt: spentAt, Comment: "кофе", CreatedAt: createdAt}, body)
	s.Contains(rec.Body.String(), `"updated_at":null`)
}

func (s *Suite) TestCreateExpense_DefaultSpentAt() {
	s.usecase.EXPECT().
		CreateExpense(gomock.Any(), entity.Expense{UserID: userA, Amount: 35000, Category: entity.CategoryFood}).
		Return(testExpense(), nil)

	rec := s.serve(http.MethodPost, "/api/expenses", `{"user_id":100,"amount":35000,"category":"food"}`)

	s.Equal(http.StatusCreated, rec.Code)
}

func (s *Suite) TestCreateExpense_InvalidJSON() {
	rec := s.serve(http.MethodPost, "/api/expenses", `{"user_id":`)

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Equal("application/json", rec.Header().Get("Content-Type"))
	s.Contains(rec.Body.String(), `"code":"invalid_request"`)
}

func (s *Suite) TestCreateExpense_FractionalAmount() {
	rec := s.serve(http.MethodPost, "/api/expenses", `{"user_id":100,"amount":350.5,"category":"food"}`)

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Contains(rec.Body.String(), `"code":"invalid_request"`)
}

func (s *Suite) TestCreateExpense_InvalidAmount() {
	s.usecase.EXPECT().
		CreateExpense(gomock.Any(), entity.Expense{UserID: userA, Amount: 0, Category: entity.CategoryFood}).
		Return(entity.Expense{}, entity.ErrInvalidAmount)

	rec := s.serve(http.MethodPost, "/api/expenses", `{"user_id":100,"amount":0,"category":"food"}`)

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Contains(rec.Body.String(), `"code":"invalid_amount"`)
}

func (s *Suite) TestCreateExpense_InvalidCategory() {
	s.usecase.EXPECT().
		CreateExpense(gomock.Any(), entity.Expense{UserID: userA, Amount: 35000, Category: "travel"}).
		Return(entity.Expense{}, entity.ErrInvalidCategory)

	rec := s.serve(http.MethodPost, "/api/expenses", `{"user_id":100,"amount":35000,"category":"travel"}`)

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Contains(rec.Body.String(), `"code":"invalid_category"`)
}

func (s *Suite) TestCreateExpense_UnknownMember() {
	s.usecase.EXPECT().
		CreateExpense(gomock.Any(), gomock.Any()).
		Return(entity.Expense{}, entity.ErrUnknownUser)

	rec := s.serve(http.MethodPost, "/api/expenses", `{"user_id":999,"amount":35000,"category":"food"}`)

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Contains(rec.Body.String(), `"code":"unknown_member"`)
}

func (s *Suite) TestCreateExpense_InternalError() {
	s.usecase.EXPECT().
		CreateExpense(gomock.Any(), gomock.Any()).
		Return(entity.Expense{}, errors.New("u.repo.CreateExpense: connection refused"))

	rec := s.serve(http.MethodPost, "/api/expenses", `{"user_id":100,"amount":35000,"category":"food"}`)

	s.Equal(http.StatusInternalServerError, rec.Code)
	s.Contains(rec.Body.String(), `"code":"internal"`)
	s.NotContains(rec.Body.String(), "connection refused")
}
