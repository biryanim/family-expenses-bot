package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestNewRouter_PathWithoutPrefix() {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/expenses", nil)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusNotFound, rec.Code)
}

func (s *Suite) TestNewRouter_CORSPreflightAllowed() {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodOptions, "/api/expenses/1", nil)
	req.Header.Set("Origin", allowedOrigin)
	req.Header.Set("Access-Control-Request-Method", http.MethodPut)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	s.Equal(allowedOrigin, rec.Header().Get("Access-Control-Allow-Origin"))
	s.Contains(rec.Header().Get("Access-Control-Allow-Methods"), http.MethodPut)
}

func (s *Suite) TestNewRouter_CORSAllowedOrigin() {
	s.usecase.EXPECT().
		ListExpenses(gomock.Any(), gomock.Any()).
		Return([]entity.Expense{}, nil)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/expenses", nil)
	req.Header.Set("Origin", allowedOrigin)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.Equal(allowedOrigin, rec.Header().Get("Access-Control-Allow-Origin"))
}

func (s *Suite) TestNewRouter_CORSForeignOrigin() {
	s.usecase.EXPECT().
		ListExpenses(gomock.Any(), gomock.Any()).
		Return([]entity.Expense{}, nil)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/expenses", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	s.Empty(rec.Header().Get("Access-Control-Allow-Origin"))
}
