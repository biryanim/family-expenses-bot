package llm_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/biryanim/family-expenses-bot/internal/agent"
	"github.com/biryanim/family-expenses-bot/internal/config"
	"github.com/biryanim/family-expenses-bot/internal/llm"
)

const (
	apiKey = "secret-test-key"
	model  = "test-model"
)

var _ agent.LLMClient = (*llm.Client)(nil)

type recordedRequest struct {
	path string
	auth string
	body map[string]any
}

type Suite struct {
	suite.Suite

	mu       sync.Mutex
	handle   http.HandlerFunc
	requests []recordedRequest

	server *httptest.Server
	client *llm.Client
}

func (s *Suite) SetupTest() {
	s.requests = nil
	s.handle = nil
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)

		s.mu.Lock()
		s.requests = append(s.requests, recordedRequest{path: r.URL.Path, auth: r.Header.Get("Authorization"), body: body})
		handle := s.handle
		s.mu.Unlock()

		handle(w, r)
	}))
	s.client = s.newClient(5 * time.Second)
}

func (s *Suite) TearDownTest() {
	s.server.Close()
}

func TestLLM(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) newClient(timeout time.Duration) *llm.Client {
	return llm.New(config.LLM{BaseURL: s.server.URL + "/v1", APIKey: apiKey, Model: model, Timeout: timeout})
}

func (s *Suite) recorded() []recordedRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]recordedRequest(nil), s.requests...)
}

func respondJSON(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After-Ms", "1")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}
}

func completionJSON(message string) string {
	return `{"id":"chatcmpl-1","object":"chat.completion","created":1,"model":"test-model","choices":[{"index":0,"finish_reason":"stop","message":` + message + `}]}`
}
