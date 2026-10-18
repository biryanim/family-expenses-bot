package http

import (
	"encoding/json"
	"net/http"

	"github.com/biryanim/family-expenses-bot/internal/transport/http/openapi"
)

func HandleRequestError(w http.ResponseWriter, _ *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(openapi.Error{Code: openapi.ErrorCodeInvalidRequest, Message: err.Error()})
}
