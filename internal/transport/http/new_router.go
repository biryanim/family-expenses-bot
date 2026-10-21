package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/biryanim/family-expenses-bot/internal/transport/http/openapi"
)

const maxRequestBodyBytes = 1 << 20

func NewRouter(s *Server, allowedOrigins []string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Content-Type"},
		ExposedHeaders:   []string{"X-Total-Count"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	strict := openapi.NewStrictHandlerWithOptions(s, nil, openapi.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  HandleRequestError,
		ResponseErrorHandlerFunc: s.HandleResponseError,
	})
	handler := openapi.HandlerWithOptions(strict, openapi.ChiServerOptions{
		BaseURL:          "/api",
		BaseRouter:       r,
		ErrorHandlerFunc: HandleRequestError,
	})
	return http.MaxBytesHandler(handler, maxRequestBodyBytes)
}
