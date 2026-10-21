package tools

import (
	"github.com/biryanim/family-expenses-bot/internal/agent"
	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (t *UpdateExpense) Definition() agent.Tool {
	categories := make([]any, 0, len(entity.Categories()))
	for _, c := range entity.Categories() {
		categories = append(categories, string(c))
	}
	return agent.Tool{
		Name:        "update_expense",
		Description: "Изменить трату пользователя по id. Меняются только переданные поля.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":       map[string]any{"type": "integer", "description": "id траты из list_expenses или get_last_expense"},
				"amount":   map[string]any{"type": "number", "description": "Новая сумма в рублях, не более двух знаков после точки"},
				"category": map[string]any{"type": "string", "enum": categories, "description": "Новый код категории"},
				"comment":  map[string]any{"type": "string", "description": "Новое описание траты; не длиннее 200 символов"},
				"date":     map[string]any{"type": "string", "description": "Новая дата траты YYYY-MM-DD"},
			},
			"required":             []any{"id"},
			"additionalProperties": false,
		},
	}
}
