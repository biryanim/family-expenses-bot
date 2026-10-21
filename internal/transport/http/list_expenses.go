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

	filter.Limit = entity.DefaultPageLimit
	if request.Params.Limit != nil {
		filter.Limit = *request.Params.Limit
	}
	if request.Params.Offset != nil {
		filter.Offset = *request.Params.Offset
	}

	page, err := s.usecase.ListExpenses(ctx, filter)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidCategory):
			return openapi.ListExpenses400JSONResponse{BadRequestJSONResponse: openapi.BadRequestJSONResponse{
				Code: openapi.ErrorCodeInvalidCategory, Message: "unknown category",
			}}, nil
		case errors.Is(err, entity.ErrInvalidPagination):
			return openapi.ListExpenses400JSONResponse{BadRequestJSONResponse: openapi.BadRequestJSONResponse{
				Code: openapi.ErrorCodeInvalidPagination, Message: "limit must be 1..500 and offset must not be negative",
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

	response := openapi.ListExpenses200JSONResponse{
		Body:    make([]openapi.Expense, 0, len(page.Items)),
		Headers: openapi.ListExpenses200ResponseHeaders{XTotalCount: page.Count},
	}
	for _, e := range page.Items {
		response.Body = append(response.Body, openapi.Expense{
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
