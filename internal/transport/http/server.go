package http

import (
	"log/slog"

	"github.com/biryanim/family-expenses-bot/internal/transport/http/openapi"
)

var _ openapi.StrictServerInterface = (*Server)(nil)

type Server struct {
	usecase Usecase
	logger  *slog.Logger
}

func New(u Usecase, logger *slog.Logger) *Server {
	return &Server{usecase: u, logger: logger}
}
