package http

import (
	"context"
	"errors"
	"log/slog"

	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/internal/transport/http/openapi"
)

func (s *Server) GetExpense(ctx context.Context, request openapi.GetExpenseRequestObject) (openapi.GetExpenseResponseObject, error) {
	expense, err := s.usecase.GetExpense(ctx, request.Id)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			return openapi.GetExpense404JSONResponse{NotFoundJSONResponse: openapi.NotFoundJSONResponse{
				Code: openapi.ErrorCodeNotFound, Message: "expense not found",
			}}, nil
		default:
			s.logger.ErrorContext(ctx, "get expense failed", slog.Int64("id", request.Id), slog.Any("error", err))
			return openapi.GetExpense500JSONResponse{InternalErrorJSONResponse: openapi.InternalErrorJSONResponse{
				Code: openapi.ErrorCodeInternal, Message: "internal server error",
			}}, nil
		}
	}

	return openapi.GetExpense200JSONResponse{
		Id:        expense.ID,
		UserId:    expense.UserID,
		Amount:    expense.Amount,
		Category:  openapi.Category(expense.Category),
		SpentAt:   expense.SpentAt,
		Comment:   expense.Comment,
		CreatedAt: expense.CreatedAt,
		UpdatedAt: expense.UpdatedAt,
	}, nil
}
