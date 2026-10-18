package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/biryanim/family-expenses-bot/internal/transport/http/openapi"
)

func (s *Server) HandleResponseError(w http.ResponseWriter, r *http.Request, err error) {
	s.logger.ErrorContext(r.Context(), "write response failed", slog.String("path", r.URL.Path), slog.Any("error", err))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(openapi.Error{Code: openapi.ErrorCodeInternal, Message: "internal server error"})
}
