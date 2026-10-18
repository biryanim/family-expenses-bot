package llm

import (
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/biryanim/family-expenses-bot/internal/config"
)

const maxRetries = 2

type Client struct {
	client  openai.Client
	model   string
	timeout time.Duration
}

func New(cfg config.LLM) *Client {
	return &Client{
		client: openai.NewClient(
			option.WithBaseURL(cfg.BaseURL),
			option.WithAPIKey(cfg.APIKey),
			option.WithMaxRetries(maxRetries),
		),
		model:   cfg.Model,
		timeout: cfg.Timeout,
	}
}
