package tools

import (
	"github.com/biryanim/family-expenses-bot/internal/agent"
	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (t *AddExpense) Definition() agent.Tool {
	categories := make([]any, 0, len(entity.Categories()))
	for _, c := range entity.Categories() {
		categories = append(categories, string(c))
	}
	return agent.Tool{
		Name:        "add_expense",
		Description: "Добавить трату пользователя.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"amount":   map[string]any{"type": "number", "description": "Сумма в рублях, не более двух знаков после точки, например 350 или 199.90"},
				"category": map[string]any{"type": "string", "enum": categories, "description": "Код категории"},
				"comment":  map[string]any{"type": "string", "description": "Что купили, 1–3 слова, например «кофе» или «такси»; не длиннее 200 символов"},
				"date":     map[string]any{"type": "string", "description": "Дата траты YYYY-MM-DD; не указывай, если трата сегодня"},
			},
			"required":             []any{"amount", "category", "comment"},
			"additionalProperties": false,
		},
	}
}
