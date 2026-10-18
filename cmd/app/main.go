// Command app runs the family expenses service: REST API and Telegram bot in one process.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/biryanim/family-expenses-bot/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		// The logger is not configured yet, so report straight to stderr.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.Log.Level}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg, logger); err != nil {
		logger.Error("app stopped with error", slog.Any("error", err))
		os.Exit(1)
	}
}

// run starts the application components and blocks until ctx is cancelled.
func run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	logger.Info("app started", slog.Any("config", cfg))

	<-ctx.Done()

	logger.Info("app stopped")
	return nil
}
