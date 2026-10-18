package http_test

import (
	"errors"
	"net/http"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestListMembers_Success() {
	s.usecase.EXPECT().
		ListMembers(gomock.Any()).
		Return([]entity.User{
			{TelegramID: 50, FirstName: "Анна", CreatedAt: createdAt},
			{TelegramID: 100, Username: "petr", FirstName: "Пётр", LastName: "Иванов", CreatedAt: createdAt, UpdatedAt: &updatedAt},
		}, nil)

	rec := s.serve(http.MethodGet, "/api/members", "")

	s.Equal(http.StatusOK, rec.Code)
	s.JSONEq(`[
		{"telegram_id":50,"username":"","first_name":"Анна","last_name":""},
		{"telegram_id":100,"username":"petr","first_name":"Пётр","last_name":"Иванов"}
	]`, rec.Body.String())
}

func (s *Suite) TestListMembers_Empty() {
	s.usecase.EXPECT().
		ListMembers(gomock.Any()).
		Return(nil, nil)

	rec := s.serve(http.MethodGet, "/api/members", "")

	s.Equal(http.StatusOK, rec.Code)
	s.JSONEq(`[]`, rec.Body.String())
}

func (s *Suite) TestListMembers_InternalError() {
	s.usecase.EXPECT().
		ListMembers(gomock.Any()).
		Return(nil, errors.New("u.repo.ListUsers: connection refused"))

	rec := s.serve(http.MethodGet, "/api/members", "")

	s.Equal(http.StatusInternalServerError, rec.Code)
	s.Contains(rec.Body.String(), `"code":"internal"`)
	s.NotContains(rec.Body.String(), "connection refused")
}
