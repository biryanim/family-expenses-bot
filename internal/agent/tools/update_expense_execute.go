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

func (t *UpdateExpense) Execute(ctx context.Context, userID int64, arguments string) (any, error) {
	var args struct {
		ID       int64        `json:"id"`
		Amount   *json.Number `json:"amount"`
		Category *string      `json:"category"`
		Comment  *string      `json:"comment"`
		Date     *string      `json:"date"`
	}
	decoder := json.NewDecoder(strings.NewReader(arguments))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&args); err != nil {
		return agent.ToolError{Error: "invalid arguments: " + err.Error()}, nil
	}

	var amount int64
	if args.Amount != nil {
		parsed, err := entity.ParseAmount(args.Amount.String())
		if err != nil {
			return agent.ToolError{Error: "invalid amount: expected rubles with at most two decimal places"}, nil
		}
		amount = parsed
	}
	var spentAt time.Time
	if args.Date != nil {
		date, err := time.ParseInLocation(time.DateOnly, *args.Date, time.Local)
		if err != nil {
			return agent.ToolError{Error: "invalid date: expected YYYY-MM-DD"}, nil
		}
		spentAt = date
	}

	expense, err := t.usecase.GetExpenseForUser(ctx, userID, args.ID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return agent.ToolError{Error: "expense not found"}, nil
		}
		return nil, fmt.Errorf("t.usecase.GetExpenseForUser: %w", err)
	}

	if args.Amount != nil {
		expense.Amount = amount
	}
	if args.Category != nil {
		expense.Category = entity.Category(*args.Category)
		if !expense.Category.Valid() {
			expense.Category = entity.CategoryOther
		}
	}
	if args.Comment != nil {
		expense.Comment = strings.TrimSpace(*args.Comment)
	}
	if args.Date != nil {
		expense.SpentAt = spentAt
	}

	updated, err := t.usecase.UpdateExpenseForUser(ctx, userID, expense)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			return agent.ToolError{Error: "expense not found"}, nil
		case errors.Is(err, entity.ErrInvalidAmount):
			return agent.ToolError{Error: "invalid amount"}, nil
		case errors.Is(err, entity.ErrInvalidCategory):
			return agent.ToolError{Error: "invalid category"}, nil
		case errors.Is(err, entity.ErrInvalidDate):
			return agent.ToolError{Error: "invalid date"}, nil
		}
		return nil, fmt.Errorf("t.usecase.UpdateExpenseForUser: %w", err)
	}

	return Expense{
		ID:       updated.ID,
		Amount:   entity.FormatAmount(updated.Amount),
		Category: updated.Category,
		Date:     updated.SpentAt.In(time.Local).Format(time.DateOnly),
		Comment:  updated.Comment,
	}, nil
}
