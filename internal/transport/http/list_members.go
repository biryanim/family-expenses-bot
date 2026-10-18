package http

import (
	"context"
	"log/slog"

	"github.com/biryanim/family-expenses-bot/internal/transport/http/openapi"
)

func (s *Server) ListMembers(ctx context.Context, _ openapi.ListMembersRequestObject) (openapi.ListMembersResponseObject, error) {
	users, err := s.usecase.ListMembers(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "list members failed", slog.Any("error", err))
		return openapi.ListMembers500JSONResponse{InternalErrorJSONResponse: openapi.InternalErrorJSONResponse{
			Code: openapi.ErrorCodeInternal, Message: "internal server error",
		}}, nil
	}

	response := make(openapi.ListMembers200JSONResponse, 0, len(users))
	for _, u := range users {
		response = append(response, openapi.Member{
			TelegramId: u.TelegramID,
			Username:   u.Username,
			FirstName:  u.FirstName,
			LastName:   u.LastName,
		})
	}
	return response, nil
}
