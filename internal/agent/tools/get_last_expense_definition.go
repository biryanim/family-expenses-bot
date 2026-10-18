package tools

import "github.com/biryanim/family-expenses-bot/internal/agent"

func (t *GetLastExpense) Definition() agent.Tool {
	return agent.Tool{
		Name:        "get_last_expense",
		Description: "Получить последнюю добавленную трату пользователя, независимо от её даты.",
		Parameters: map[string]any{
			"type":                 "object",
			"properties":           map[string]any{},
			"additionalProperties": false,
		},
	}
}
