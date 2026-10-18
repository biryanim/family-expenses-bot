package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/biryanim/family-expenses-bot/internal/config"
	"github.com/biryanim/family-expenses-bot/internal/database/postgres"
	"github.com/biryanim/family-expenses-bot/internal/database/repository"
	httpapi "github.com/biryanim/family-expenses-bot/internal/transport/http"
	"github.com/biryanim/family-expenses-bot/internal/usecase"
)

const shutdownTimeout = 10 * time.Second

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.Log.Level}))
	slog.SetDefault(logger)

	if err := run(cfg, logger); err != nil {
		logger.Error("app failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(cfg config.Config, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.Connect(ctx, logger, cfg.DB.DSN)
	if err != nil {
		return err
	}
	defer pool.Close()

	uc := usecase.New(repository.New(pool))
	srv := &http.Server{
		Handler:           httpapi.NewRouter(httpapi.New(uc, logger), cfg.HTTP.AllowedOrigins),
		ReadHeaderTimeout: 5 * time.Second,
	}

	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", ":"+strconv.Itoa(cfg.HTTP.Port))
	if err != nil {
		return fmt.Errorf("lc.Listen: %w", err)
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(ln)
	}()

	logger.Info("http server started", slog.String("addr", ln.Addr().String()))
	logger.Info("app started", slog.Any("config", cfg))

	select {
	case err := <-serveErr:
		return fmt.Errorf("srv.Serve: %w", err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("srv.Shutdown: %w", err)
	}
	if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("srv.Serve: %w", err)
	}

	logger.Info("app stopped")
	return nil
}
