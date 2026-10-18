package http

import (
	"context"
	"errors"
	"log/slog"

	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/internal/transport/http/openapi"
)

func (s *Server) ListExpenses(ctx context.Context, request openapi.ListExpensesRequestObject) (openapi.ListExpensesResponseObject, error) {
	var filter entity.ExpenseFilter
	filter.UserID = request.Params.UserId
	if request.Params.From != nil {
		filter.From = *request.Params.From
	}
	if request.Params.To != nil {
		filter.To = *request.Params.To
	}
	if request.Params.Category != nil {
		filter.Category = entity.Category(*request.Params.Category)
	}

	expenses, err := s.usecase.ListExpenses(ctx, filter)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidCategory):
			return openapi.ListExpenses400JSONResponse{BadRequestJSONResponse: openapi.BadRequestJSONResponse{
				Code: openapi.ErrorCodeInvalidCategory, Message: "unknown category",
			}}, nil
		case errors.Is(err, entity.ErrInvalidDate):
			return openapi.ListExpenses400JSONResponse{BadRequestJSONResponse: openapi.BadRequestJSONResponse{
				Code: openapi.ErrorCodeInvalidDate, Message: "from must not be later than to",
			}}, nil
		default:
			s.logger.ErrorContext(ctx, "list expenses failed", slog.Any("error", err))
			return openapi.ListExpenses500JSONResponse{InternalErrorJSONResponse: openapi.InternalErrorJSONResponse{
				Code: openapi.ErrorCodeInternal, Message: "internal server error",
			}}, nil
		}
	}

	response := make(openapi.ListExpenses200JSONResponse, 0, len(expenses))
	for _, e := range expenses {
		response = append(response, openapi.Expense{
			Id:        e.ID,
			UserId:    e.UserID,
			Amount:    e.Amount,
			Category:  openapi.Category(e.Category),
			SpentAt:   e.SpentAt,
			Comment:   e.Comment,
			CreatedAt: e.CreatedAt,
			UpdatedAt: e.UpdatedAt,
		})
	}
	return response, nil
}
