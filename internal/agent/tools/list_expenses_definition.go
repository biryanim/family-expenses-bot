package tools

import (
	"github.com/biryanim/family-expenses-bot/internal/agent"
	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (t *ListExpenses) Definition() agent.Tool {
	categories := make([]any, 0, len(entity.Categories()))
	for _, c := range entity.Categories() {
		categories = append(categories, string(c))
	}
	return agent.Tool{
		Name:        "list_expenses",
		Description: "Показать траты пользователя с фильтрами. Возвращает количество, итоговую сумму total и до 30 самых поздних трат.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"from":     map[string]any{"type": "string", "description": "Начало периода YYYY-MM-DD включительно"},
				"to":       map[string]any{"type": "string", "description": "Конец периода YYYY-MM-DD включительно"},
				"category": map[string]any{"type": "string", "enum": categories, "description": "Код категории"},
			},
			"additionalProperties": false,
		},
	}
}
