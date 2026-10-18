package http_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
	httpapi "github.com/biryanim/family-expenses-bot/internal/transport/http"
	"github.com/biryanim/family-expenses-bot/mock"
)

const (
	specPath      = "../../../api/openapi.yaml"
	allowedOrigin = "http://localhost:5173"
	userA         = int64(100)
)

var (
	spentAt   = time.Date(2026, 10, 17, 9, 30, 0, 0, time.UTC)
	createdAt = time.Date(2026, 10, 17, 9, 31, 0, 0, time.UTC)
	updatedAt = time.Date(2026, 10, 18, 7, 0, 0, 0, time.UTC)
	from      = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to        = time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
)

type Suite struct {
	suite.Suite

	ctrl *gomock.Controller

	usecase *mock.MockUsecase
	router  http.Handler

	specRouter routers.Router
}

func (s *Suite) SetupSuite() {
	doc, err := openapi3.NewLoader().LoadFromFile(specPath)
	s.Require().NoError(err)
	s.specRouter, err = legacy.NewRouter(doc)
	s.Require().NoError(err)
}

func (s *Suite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.usecase = mock.NewMockUsecase(s.ctrl)
	s.router = httpapi.NewRouter(httpapi.New(s.usecase, slog.New(slog.DiscardHandler)), []string{allowedOrigin})
}

func (s *Suite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *Suite) SetupSubTest() {
	s.SetupTest()
}

func (s *Suite) TearDownSubTest() {
	s.TearDownTest()
}

func TestHTTP(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) serve(method, target, body string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequestWithContext(context.Background(), method, target, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	route, pathParams, err := s.specRouter.FindRoute(req)
	s.Require().NoError(err)
	err = openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
		},
		Status:  rec.Code,
		Header:  rec.Header(),
		Body:    io.NopCloser(strings.NewReader(rec.Body.String())),
		Options: &openapi3filter.Options{IncludeResponseStatus: true},
	})
	s.Require().NoError(err, "response does not match spec: %d %s", rec.Code, rec.Body.String())

	return rec
}

func testExpense() entity.Expense {
	return entity.Expense{
		ID:        7,
		UserID:    userA,
		Amount:    35000,
		Category:  entity.CategoryFood,
		SpentAt:   spentAt,
		Comment:   "кофе",
		CreatedAt: createdAt,
	}
}
