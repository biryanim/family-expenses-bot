package http

import (
	"context"
	"errors"
	"log/slog"

	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/internal/transport/http/openapi"
)

func (s *Server) DeleteExpense(ctx context.Context, request openapi.DeleteExpenseRequestObject) (openapi.DeleteExpenseResponseObject, error) {
	if err := s.usecase.DeleteExpense(ctx, request.Id); err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			return openapi.DeleteExpense404JSONResponse{NotFoundJSONResponse: openapi.NotFoundJSONResponse{
				Code: openapi.ErrorCodeNotFound, Message: "expense not found",
			}}, nil
		default:
			s.logger.ErrorContext(ctx, "delete expense failed", slog.Int64("id", request.Id), slog.Any("error", err))
			return openapi.DeleteExpense500JSONResponse{InternalErrorJSONResponse: openapi.InternalErrorJSONResponse{
				Code: openapi.ErrorCodeInternal, Message: "internal server error",
			}}, nil
		}
	}
	return openapi.DeleteExpense204Response{}, nil
}
