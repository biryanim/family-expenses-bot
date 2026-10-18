package repository

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestNewPoolErrorHidesPassword(t *testing.T) {
	const password = "s3cr3t-pass"
	tests := []struct {
		name string
		url  string
	}{
		{name: "unreachable host", url: "postgres://expenses:" + password + "@127.0.0.1:1/expenses?sslmode=disable"},
		{name: "malformed url", url: "postgres://expenses:" + password + "@127.0.0.1:notaport/expenses"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			pool, err := NewPool(ctx, tt.url)
			if err == nil {
				pool.Close()
				t.Fatal("NewPool() error = nil, want error")
			}
			if strings.Contains(err.Error(), password) {
				t.Errorf("NewPool() error contains password: %v", err)
			}
		})
	}
}
