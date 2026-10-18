package http_test

import (
	"errors"
	"net/http"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestDeleteExpense_Success() {
	s.usecase.EXPECT().
		DeleteExpense(gomock.Any(), int64(7)).
		Return(nil)

	rec := s.serve(http.MethodDelete, "/api/expenses/7", "")

	s.Equal(http.StatusNoContent, rec.Code)
	s.Empty(rec.Body.String())
}

func (s *Suite) TestDeleteExpense_InvalidID() {
	rec := s.serve(http.MethodDelete, "/api/expenses/abc", "")

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Contains(rec.Body.String(), `"code":"invalid_request"`)
}

func (s *Suite) TestDeleteExpense_NotFound() {
	s.usecase.EXPECT().
		DeleteExpense(gomock.Any(), int64(999)).
		Return(entity.ErrNotFound)

	rec := s.serve(http.MethodDelete, "/api/expenses/999", "")

	s.Equal(http.StatusNotFound, rec.Code)
	s.Contains(rec.Body.String(), `"code":"not_found"`)
}

func (s *Suite) TestDeleteExpense_InternalError() {
	s.usecase.EXPECT().
		DeleteExpense(gomock.Any(), int64(7)).
		Return(errors.New("u.repo.DeleteExpense: connection refused"))

	rec := s.serve(http.MethodDelete, "/api/expenses/7", "")

	s.Equal(http.StatusInternalServerError, rec.Code)
	s.Contains(rec.Body.String(), `"code":"internal"`)
	s.NotContains(rec.Body.String(), "connection refused")
}
