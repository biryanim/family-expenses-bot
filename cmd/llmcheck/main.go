package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/biryanim/family-expenses-bot/internal/agent"
	"github.com/biryanim/family-expenses-bot/internal/config"
	"github.com/biryanim/family-expenses-bot/internal/llm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("base_url: %s\nmodel: %s\ntimeout: %s\n", cfg.LLM.BaseURL, cfg.LLM.Model, cfg.LLM.Timeout)

	messages := []agent.Message{
		{Role: agent.RoleSystem, Content: "Ты помощник учёта семейных расходов. Когда пользователь сообщает о трате, вызови инструмент add_expense. Сумма в рублях."},
		{Role: agent.RoleUser, Content: "кофе 350"},
	}
	tools := []agent.Tool{{
		Name:        "add_expense",
		Description: "Добавить трату пользователя",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"amount":      map[string]any{"type": "number", "description": "Сумма в рублях"},
				"category":    map[string]any{"type": "string", "description": "Категория траты"},
				"description": map[string]any{"type": "string", "description": "Описание траты"},
			},
			"required": []any{"amount", "category", "description"},
		},
	}}

	msg, err := llm.New(cfg.LLM).CreateChatCompletion(context.Background(), messages, tools)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if msg.Content != "" {
		fmt.Printf("content: %s\n", msg.Content)
	}
	valid := 0
	for _, tc := range msg.ToolCalls {
		fmt.Printf("tool_call: id=%s name=%s arguments=%s\n", tc.ID, tc.Name, tc.Arguments)
		if json.Valid([]byte(tc.Arguments)) {
			valid++
		}
	}
	if valid == 0 {
		fmt.Fprintln(os.Stderr, "no tool call with valid JSON arguments")
		os.Exit(1)
	}
}
