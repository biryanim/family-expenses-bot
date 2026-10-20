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

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/biryanim/family-expenses-bot/internal/agent"
	"github.com/biryanim/family-expenses-bot/internal/agent/history"
	"github.com/biryanim/family-expenses-bot/internal/agent/tools"
	"github.com/biryanim/family-expenses-bot/internal/config"
	"github.com/biryanim/family-expenses-bot/internal/database/postgres"
	"github.com/biryanim/family-expenses-bot/internal/database/repository"
	"github.com/biryanim/family-expenses-bot/internal/llm"
	httpapi "github.com/biryanim/family-expenses-bot/internal/transport/http"
	"github.com/biryanim/family-expenses-bot/internal/transport/telegram"
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
	expenseAgent := agent.New(llm.New(cfg.LLM), history.New(),
		tools.NewAddExpense(uc),
		tools.NewListExpenses(uc),
		tools.NewGetLastExpense(uc),
		tools.NewUpdateExpense(uc),
		tools.NewDeleteExpense(uc),
	)

	b, err := bot.New(cfg.Telegram.Token,
		bot.WithErrorsHandler(func(err error) {
			logger.Error("telegram bot error", slog.Any("error", err))
		}),
		bot.WithAllowedUpdates(bot.AllowedUpdates{"message"}),
	)
	if err != nil {
		return fmt.Errorf("bot.New: %w", err)
	}
	b.RegisterHandlerMatchFunc(func(*models.Update) bool { return true }, telegram.New(uc, expenseAgent, b, logger).HandleUpdate)

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

	botDone := make(chan struct{})
	go func() {
		b.Start(ctx)
		close(botDone)
	}()
	logger.Info("telegram bot started")
	logger.Info("app started", slog.Any("config", cfg), slog.String("timezone", time.Local.String()))

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
	<-botDone

	logger.Info("app stopped")
	return nil
}
