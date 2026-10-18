package config

import (
	"bytes"
	"log/slog"
	"reflect"
	"strings"
	"testing"
)

func validEnv() map[string]string {
	return map[string]string{
		envTelegramBotToken: "secret-bot-token",
		envLLMBaseURL:       "https://llm.example.com/v1",
		envLLMAPIKey:        "secret-llm-key",
		envLLMModel:         "test-model",
		envDatabaseDSN:      "postgres://user:secret-db-pass@db:5432/expenses",
	}
}

func getenvFrom(env map[string]string) func(string) string {
	return func(name string) string { return env[name] }
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		patch   map[string]string
		wantErr []string
		check   func(t *testing.T, cfg Config)
	}{
		{
			name: "all values set",
			patch: map[string]string{
				envHTTPPort:    "9090",
				envLogLevel:    "debug",
				envCORSOrigins: "https://panel.example.com",
			},
			check: func(t *testing.T, cfg Config) {
				want := Config{
					Telegram: Telegram{Token: "secret-bot-token"},
					LLM:      LLM{BaseURL: "https://llm.example.com/v1", APIKey: "secret-llm-key", Model: "test-model"},
					DB:       DB{DSN: "postgres://user:secret-db-pass@db:5432/expenses"},
					HTTP:     HTTP{Port: 9090, AllowedOrigins: []string{"https://panel.example.com"}},
					Log:      Log{Level: slog.LevelDebug},
				}
				if !reflect.DeepEqual(cfg, want) {
					t.Errorf("got %+v, want %+v", cfg, want)
				}
			},
		},
		{
			name: "defaults for port and level",
			check: func(t *testing.T, cfg Config) {
				if cfg.HTTP.Port != 8080 {
					t.Errorf("port = %d, want 8080", cfg.HTTP.Port)
				}
				if cfg.Log.Level != slog.LevelInfo {
					t.Errorf("level = %v, want info", cfg.Log.Level)
				}
				if want := []string{"http://localhost:5173"}; !reflect.DeepEqual(cfg.HTTP.AllowedOrigins, want) {
					t.Errorf("allowed origins = %v, want %v", cfg.HTTP.AllowedOrigins, want)
				}
			},
		},
		{
			name:  "blank cors origins fall back to default",
			patch: map[string]string{envCORSOrigins: " , "},
			check: func(t *testing.T, cfg Config) {
				if want := []string{"http://localhost:5173"}; !reflect.DeepEqual(cfg.HTTP.AllowedOrigins, want) {
					t.Errorf("allowed origins = %v, want %v", cfg.HTTP.AllowedOrigins, want)
				}
			},
		},
		{
			name:  "several cors origins with spaces and empty items",
			patch: map[string]string{envCORSOrigins: "http://localhost:5173, http://192.168.1.10:3000,,"},
			check: func(t *testing.T, cfg Config) {
				want := []string{"http://localhost:5173", "http://192.168.1.10:3000"}
				if !reflect.DeepEqual(cfg.HTTP.AllowedOrigins, want) {
					t.Errorf("allowed origins = %v, want %v", cfg.HTTP.AllowedOrigins, want)
				}
			},
		},
		{
			name:    "one required variable missing",
			patch:   map[string]string{envTelegramBotToken: ""},
			wantErr: []string{envTelegramBotToken},
		},
		{
			name:    "several required variables missing",
			patch:   map[string]string{envLLMAPIKey: "", envDatabaseDSN: ""},
			wantErr: []string{envLLMAPIKey, envDatabaseDSN},
		},
		{
			name:    "whitespace-only value is treated as missing",
			patch:   map[string]string{envLLMModel: "   "},
			wantErr: []string{envLLMModel},
		},
		{
			name:    "non-numeric port",
			patch:   map[string]string{envHTTPPort: "abc"},
			wantErr: []string{envHTTPPort, `"abc"`},
		},
		{
			name:    "zero port",
			patch:   map[string]string{envHTTPPort: "0"},
			wantErr: []string{envHTTPPort, `"0"`},
		},
		{
			name:    "port out of range",
			patch:   map[string]string{envHTTPPort: "70000"},
			wantErr: []string{envHTTPPort, `"70000"`},
		},
		{
			name:    "unknown log level",
			patch:   map[string]string{envLogLevel: "verbose"},
			wantErr: []string{envLogLevel, `"verbose"`},
		},
		{
			name:  "log level is case-insensitive",
			patch: map[string]string{envLogLevel: "WARN"},
			check: func(t *testing.T, cfg Config) {
				if cfg.Log.Level != slog.LevelWarn {
					t.Errorf("level = %v, want warn", cfg.Log.Level)
				}
			},
		},
		{
			name: "missing and invalid values reported together",
			patch: map[string]string{
				envTelegramBotToken: "",
				envHTTPPort:         "abc",
				envLogLevel:         "verbose",
			},
			wantErr: []string{envTelegramBotToken, envHTTPPort, envLogLevel},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := validEnv()
			for k, v := range tt.patch {
				env[k] = v
			}

			cfg, err := load(getenvFrom(env))

			if len(tt.wantErr) > 0 {
				if err == nil {
					t.Fatalf("expected error, got nil (cfg %+v)", cfg)
				}
				for _, s := range tt.wantErr {
					if !strings.Contains(err.Error(), s) {
						t.Errorf("error %q does not contain %q", err, s)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestLoadReadsProcessEnv(t *testing.T) {
	for k, v := range validEnv() {
		t.Setenv(k, v)
	}
	t.Setenv(envHTTPPort, "")
	t.Setenv(envLogLevel, "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Telegram.Token != "secret-bot-token" {
		t.Errorf("token = %q, want value from env", cfg.Telegram.Token)
	}
}

func TestErrorDoesNotLeakSecrets(t *testing.T) {
	env := validEnv()
	env[envHTTPPort] = "abc"

	_, err := load(getenvFrom(env))
	if err == nil {
		t.Fatal("expected error")
	}
	for _, secret := range []string{"secret-bot-token", "secret-llm-key", "secret-db-pass"} {
		if strings.Contains(err.Error(), secret) {
			t.Errorf("error %q leaks secret %q", err, secret)
		}
	}
}

func TestLogValueHidesSecrets(t *testing.T) {
	cfg, err := load(getenvFrom(validEnv()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	slog.New(slog.NewJSONHandler(&buf, nil)).Info("start", slog.Any("config", cfg))
	out := buf.String()

	for _, secret := range []string{"secret-bot-token", "secret-llm-key", "secret-db-pass"} {
		if strings.Contains(out, secret) {
			t.Errorf("log output leaks secret %q: %s", secret, out)
		}
	}
	for _, visible := range []string{"https://llm.example.com/v1", "test-model", "8080", "http://localhost:5173"} {
		if !strings.Contains(out, visible) {
			t.Errorf("log output lacks %q: %s", visible, out)
		}
	}
}
