package http

import (
	"context"
	"errors"
	"log/slog"

	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/internal/transport/http/openapi"
)

func (s *Server) UpdateExpense(ctx context.Context, request openapi.UpdateExpenseRequestObject) (openapi.UpdateExpenseResponseObject, error) {
	category := entity.Category(request.Body.Category)
	comment := ""
	if request.Body.Comment != nil {
		comment = *request.Body.Comment
	}
	patch := entity.ExpensePatch{
		Amount:   &request.Body.Amount,
		Category: &category,
		SpentAt:  &request.Body.SpentAt,
		Comment:  &comment,
	}

	updated, err := s.usecase.UpdateExpense(ctx, request.Id, patch)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidAmount):
			return openapi.UpdateExpense400JSONResponse{BadRequestJSONResponse: openapi.BadRequestJSONResponse{
				Code: openapi.ErrorCodeInvalidAmount, Message: "amount must be a positive number of kopecks",
			}}, nil
		case errors.Is(err, entity.ErrInvalidCategory):
			return openapi.UpdateExpense400JSONResponse{BadRequestJSONResponse: openapi.BadRequestJSONResponse{
				Code: openapi.ErrorCodeInvalidCategory, Message: "unknown category",
			}}, nil
		case errors.Is(err, entity.ErrInvalidComment):
			return openapi.UpdateExpense400JSONResponse{BadRequestJSONResponse: openapi.BadRequestJSONResponse{
				Code: openapi.ErrorCodeInvalidComment, Message: "comment must be at most 200 characters",
			}}, nil
		case errors.Is(err, entity.ErrInvalidDate):
			return openapi.UpdateExpense400JSONResponse{BadRequestJSONResponse: openapi.BadRequestJSONResponse{
				Code: openapi.ErrorCodeInvalidDate, Message: "spent_at is required",
			}}, nil
		case errors.Is(err, entity.ErrNotFound):
			return openapi.UpdateExpense404JSONResponse{NotFoundJSONResponse: openapi.NotFoundJSONResponse{
				Code: openapi.ErrorCodeNotFound, Message: "expense not found",
			}}, nil
		default:
			s.logger.ErrorContext(ctx, "update expense failed", slog.Int64("id", request.Id), slog.Any("error", err))
			return openapi.UpdateExpense500JSONResponse{InternalErrorJSONResponse: openapi.InternalErrorJSONResponse{
				Code: openapi.ErrorCodeInternal, Message: "internal server error",
			}}, nil
		}
	}

	return openapi.UpdateExpense200JSONResponse{
		Id:        updated.ID,
		UserId:    updated.UserID,
		Amount:    updated.Amount,
		Category:  openapi.Category(updated.Category),
		SpentAt:   updated.SpentAt,
		Comment:   updated.Comment,
		CreatedAt: updated.CreatedAt,
		UpdatedAt: updated.UpdatedAt,
	}, nil
}
