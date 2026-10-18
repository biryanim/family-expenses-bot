package tools

import "github.com/biryanim/family-expenses-bot/internal/agent"

func (t *DeleteExpense) Definition() agent.Tool {
	return agent.Tool{
		Name:        "delete_expense",
		Description: "Удалить трату пользователя по id. Возвращает удалённую трату.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "integer", "description": "id траты из list_expenses или get_last_expense"},
			},
			"required":             []any{"id"},
			"additionalProperties": false,
		},
	}
}
