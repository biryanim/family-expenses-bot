package http

import (
	"context"
	"errors"
	"log/slog"

	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/internal/transport/http/openapi"
)

func (s *Server) CreateExpense(ctx context.Context, request openapi.CreateExpenseRequestObject) (openapi.CreateExpenseResponseObject, error) {
	input := entity.Expense{
		UserID:   request.Body.UserId,
		Amount:   request.Body.Amount,
		Category: entity.Category(request.Body.Category),
	}
	if request.Body.SpentAt != nil {
		input.SpentAt = *request.Body.SpentAt
	}
	if request.Body.Comment != nil {
		input.Comment = *request.Body.Comment
	}

	created, err := s.usecase.CreateExpense(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidAmount):
			return openapi.CreateExpense400JSONResponse{BadRequestJSONResponse: openapi.BadRequestJSONResponse{
				Code: openapi.ErrorCodeInvalidAmount, Message: "amount must be a positive number of kopecks",
			}}, nil
		case errors.Is(err, entity.ErrInvalidCategory):
			return openapi.CreateExpense400JSONResponse{BadRequestJSONResponse: openapi.BadRequestJSONResponse{
				Code: openapi.ErrorCodeInvalidCategory, Message: "unknown category",
			}}, nil
		case errors.Is(err, entity.ErrUnknownUser):
			return openapi.CreateExpense400JSONResponse{BadRequestJSONResponse: openapi.BadRequestJSONResponse{
				Code: openapi.ErrorCodeUnknownMember, Message: "unknown family member",
			}}, nil
		default:
			s.logger.ErrorContext(ctx, "create expense failed", slog.Any("error", err))
			return openapi.CreateExpense500JSONResponse{InternalErrorJSONResponse: openapi.InternalErrorJSONResponse{
				Code: openapi.ErrorCodeInternal, Message: "internal server error",
			}}, nil
		}
	}

	return openapi.CreateExpense201JSONResponse{
		Id:        created.ID,
		UserId:    created.UserID,
		Amount:    created.Amount,
		Category:  openapi.Category(created.Category),
		SpentAt:   created.SpentAt,
		Comment:   created.Comment,
		CreatedAt: created.CreatedAt,
		UpdatedAt: created.UpdatedAt,
	}, nil
}
