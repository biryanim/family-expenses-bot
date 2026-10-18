package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/biryanim/family-expenses-bot/internal/transport/http/openapi"
)

func NewRouter(s *Server, allowedOrigins []string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	strict := openapi.NewStrictHandlerWithOptions(s, nil, openapi.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  HandleRequestError,
		ResponseErrorHandlerFunc: s.HandleResponseError,
	})
	return openapi.HandlerWithOptions(strict, openapi.ChiServerOptions{
		BaseURL:          "/api",
		BaseRouter:       r,
		ErrorHandlerFunc: HandleRequestError,
	})
}
