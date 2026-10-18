package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/biryanim/family-expenses-bot/internal/config"
	"github.com/biryanim/family-expenses-bot/internal/repository"
)

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

	pool, err := repository.NewPool(ctx, cfg.DB.DSN)
	if err != nil {
		return err
	}
	defer pool.Close()
	logger.Info("database connected")

	logger.Info("app started", slog.Any("config", cfg))

	<-ctx.Done()

	logger.Info("app stopped")
	return nil
}
