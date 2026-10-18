// Package config loads application configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

const (
	envTelegramBotToken = "TELEGRAM_BOT_TOKEN"
	envLLMBaseURL       = "LLM_BASE_URL"
	envLLMAPIKey        = "LLM_API_KEY"
	envLLMModel         = "LLM_MODEL"
	envDatabaseURL      = "DATABASE_URL"
	envHTTPPort         = "HTTP_PORT"
	envLogLevel         = "LOG_LEVEL"

	defaultHTTPPort = 8080
)

// Config is the full application configuration.
type Config struct {
	Telegram Telegram
	LLM      LLM
	DB       DB
	HTTP     HTTP
	Log      Log
}

// Telegram holds Telegram bot settings.
type Telegram struct {
	Token string
}

// LLM holds settings for the OpenAI-compatible Chat Completions API.
type LLM struct {
	BaseURL string
	APIKey  string
	Model   string
}

// DB holds PostgreSQL settings.
type DB struct {
	URL string
}

// HTTP holds REST API server settings.
type HTTP struct {
	Port int
}

// Log holds logging settings.
type Log struct {
	Level slog.Level
}

// Load reads the configuration from the process environment.
func Load() (Config, error) {
	return load(os.Getenv)
}

func load(getenv func(string) string) (Config, error) {
	var (
		cfg     Config
		missing []string
		errs    []error
	)

	required := func(name string) string {
		v := strings.TrimSpace(getenv(name))
		if v == "" {
			missing = append(missing, name)
		}
		return v
	}

	cfg.Telegram.Token = required(envTelegramBotToken)
	cfg.LLM.BaseURL = required(envLLMBaseURL)
	cfg.LLM.APIKey = required(envLLMAPIKey)
	cfg.LLM.Model = required(envLLMModel)
	cfg.DB.URL = required(envDatabaseURL)

	if len(missing) > 0 {
		errs = append(errs, fmt.Errorf("missing required env: %s", strings.Join(missing, ", ")))
	}

	port, err := parsePort(strings.TrimSpace(getenv(envHTTPPort)))
	if err != nil {
		errs = append(errs, err)
	}
	cfg.HTTP.Port = port

	level, err := parseLevel(strings.TrimSpace(getenv(envLogLevel)))
	if err != nil {
		errs = append(errs, err)
	}
	cfg.Log.Level = level

	if err := errors.Join(errs...); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}
	return cfg, nil
}

func parsePort(raw string) (int, error) {
	if raw == "" {
		return defaultHTTPPort, nil
	}
	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("%s: invalid value %q: must be an integer in 1..65535", envHTTPPort, raw)
	}
	return port, nil
}

func parseLevel(raw string) (slog.Level, error) {
	switch strings.ToLower(raw) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("%s: invalid value %q: must be one of debug, info, warn, error", envLogLevel, raw)
	}
}

// LogValue implements slog.LogValuer and exposes only non-secret fields,
// so logging the whole config never leaks tokens or credentials.
func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("http_port", c.HTTP.Port),
		slog.String("log_level", c.Log.Level.String()),
		slog.String("llm_base_url", c.LLM.BaseURL),
		slog.String("llm_model", c.LLM.Model),
	)
}
