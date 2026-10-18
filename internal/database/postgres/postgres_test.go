package postgres

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestConnectErrorHidesPassword(t *testing.T) {
	const password = "s3cr3t-pass"
	tests := []struct {
		name string
		url  string
	}{
		{name: "unreachable host", url: "postgres://expenses:" + password + "@127.0.0.1:1/expenses?sslmode=disable"},
		{name: "malformed url", url: "postgres://expenses:" + password + "@127.0.0.1:notaport/expenses"},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			pool, err := Connect(ctx, logger, tt.url)
			if err == nil {
				pool.Close()
				t.Fatal("Connect() error = nil, want error")
			}
			if strings.Contains(err.Error(), password) {
				t.Errorf("Connect() error contains password: %v", err)
			}
		})
	}
}
