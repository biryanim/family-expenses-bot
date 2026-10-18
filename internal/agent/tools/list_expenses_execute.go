package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/biryanim/family-expenses-bot/internal/agent"
	"github.com/biryanim/family-expenses-bot/internal/entity"
)

const maxListedExpenses = 30

func (t *ListExpenses) Execute(ctx context.Context, userID int64, arguments string) (any, error) {
	var args struct {
		From     string `json:"from"`
		To       string `json:"to"`
		Category string `json:"category"`
	}
	decoder := json.NewDecoder(strings.NewReader(arguments))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&args); err != nil {
		return agent.ToolError{Error: "invalid arguments: " + err.Error()}, nil
	}

	filter := entity.ExpenseFilter{Category: entity.Category(args.Category)}
	if args.From != "" {
		from, err := time.ParseInLocation(time.DateOnly, args.From, time.Local)
		if err != nil {
			return agent.ToolError{Error: "invalid from: expected YYYY-MM-DD"}, nil
		}
		filter.From = from
	}
	if args.To != "" {
		to, err := time.ParseInLocation(time.DateOnly, args.To, time.Local)
		if err != nil {
			return agent.ToolError{Error: "invalid to: expected YYYY-MM-DD"}, nil
		}
		filter.To = to.AddDate(0, 0, 1)
	}

	expenses, err := t.usecase.ListExpensesForUser(ctx, userID, filter)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidCategory):
			return agent.ToolError{Error: "invalid category"}, nil
		case errors.Is(err, entity.ErrInvalidDate):
			return agent.ToolError{Error: "invalid period: from is after to"}, nil
		}
		return nil, fmt.Errorf("t.usecase.ListExpensesForUser: %w", err)
	}

	var total int64
	for _, e := range expenses {
		total += e.Amount
	}
	result := ExpenseList{
		Count:     len(expenses),
		Total:     entity.FormatAmount(total),
		Expenses:  []Expense{},
		Truncated: len(expenses) > maxListedExpenses,
	}
	for _, e := range expenses[:min(len(expenses), maxListedExpenses)] {
		result.Expenses = append(result.Expenses, Expense{
			ID:       e.ID,
			Amount:   entity.FormatAmount(e.Amount),
			Category: e.Category,
			Date:     e.SpentAt.In(time.Local).Format(time.DateOnly),
			Comment:  e.Comment,
		})
	}
	return result, nil
}
