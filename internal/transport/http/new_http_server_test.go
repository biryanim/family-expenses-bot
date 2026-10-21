package http_test

import (
	"time"

	httpapi "github.com/biryanim/family-expenses-bot/internal/transport/http"
)

func (s *Suite) TestNewHTTPServer_Timeouts() {
	srv := httpapi.NewHTTPServer(s.router)

	s.Equal(5*time.Second, srv.ReadHeaderTimeout)
	s.Equal(10*time.Second, srv.ReadTimeout)
	s.Equal(30*time.Second, srv.WriteTimeout)
	s.Equal(60*time.Second, srv.IdleTimeout)
	s.NotNil(srv.Handler)
}
