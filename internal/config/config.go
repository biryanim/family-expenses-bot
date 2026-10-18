package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	envTelegramBotToken = "TELEGRAM_BOT_TOKEN"
	envLLMBaseURL       = "LLM_BASE_URL"
	envLLMAPIKey        = "LLM_API_KEY"
	envLLMModel         = "LLM_MODEL"
	envLLMTimeout       = "LLM_TIMEOUT"
	envDatabaseDSN      = "DATABASE_DSN"
	envHTTPPort         = "HTTP_PORT"
	envLogLevel         = "LOG_LEVEL"
	envCORSOrigins      = "CORS_ALLOWED_ORIGINS"

	defaultHTTPPort    = 8080
	defaultCORSOrigins = "http://localhost:5173"
	defaultLLMTimeout  = 30 * time.Second
)

type Config struct {
	Telegram Telegram
	LLM      LLM
	DB       DB
	HTTP     HTTP
	Log      Log
}

type Telegram struct {
	Token string
}

type LLM struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
}

type DB struct {
	DSN string
}

type HTTP struct {
	Port           int
	AllowedOrigins []string
}

type Log struct {
	Level slog.Level
}

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
	cfg.DB.DSN = required(envDatabaseDSN)

	if len(missing) > 0 {
		errs = append(errs, fmt.Errorf("missing required env: %s", strings.Join(missing, ", ")))
	}

	port, err := parsePort(strings.TrimSpace(getenv(envHTTPPort)))
	if err != nil {
		errs = append(errs, err)
	}
	cfg.HTTP.Port = port
	cfg.HTTP.AllowedOrigins = parseOrigins(getenv(envCORSOrigins))

	timeout, err := parseTimeout(strings.TrimSpace(getenv(envLLMTimeout)))
	if err != nil {
		errs = append(errs, err)
	}
	cfg.LLM.Timeout = timeout

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

func parseOrigins(raw string) []string {
	var origins []string
	for _, origin := range strings.Split(raw, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}
	if len(origins) == 0 {
		return []string{defaultCORSOrigins}
	}
	return origins
}

func parseTimeout(raw string) (time.Duration, error) {
	if raw == "" {
		return defaultLLMTimeout, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("%s: invalid value %q: must be a positive duration like 30s", envLLMTimeout, raw)
	}
	return d, nil
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

func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("http_port", c.HTTP.Port),
		slog.String("log_level", c.Log.Level.String()),
		slog.String("llm_base_url", c.LLM.BaseURL),
		slog.String("llm_model", c.LLM.Model),
		slog.String("llm_timeout", c.LLM.Timeout.String()),
		slog.Any("cors_allowed_origins", c.HTTP.AllowedOrigins),
	)
}
