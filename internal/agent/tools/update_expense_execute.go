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

	var patch entity.ExpensePatch
	if args.Amount != nil {
		amount, err := entity.ParseAmount(args.Amount.String())
		if err != nil {
			return agent.ToolError{Error: "invalid amount: expected rubles with at most two decimal places"}, nil
		}
		patch.Amount = &amount
	}
	if args.Category != nil {
		category := entity.Category(*args.Category)
		if !category.Valid() {
			category = entity.CategoryOther
		}
		patch.Category = &category
	}
	if args.Comment != nil {
		comment := strings.TrimSpace(*args.Comment)
		patch.Comment = &comment
	}
	if args.Date != nil {
		date, err := time.ParseInLocation(time.DateOnly, *args.Date, time.Local)
		if err != nil {
			return agent.ToolError{Error: "invalid date: expected YYYY-MM-DD"}, nil
		}
		patch.SpentAt = &date
	}
	if patch.Amount == nil && patch.Category == nil && patch.Comment == nil && patch.SpentAt == nil {
		return agent.ToolError{Error: "nothing to update: pass amount, category, comment or date"}, nil
	}

	updated, err := t.usecase.UpdateExpenseForUser(ctx, userID, args.ID, patch)
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
		case errors.Is(err, entity.ErrInvalidComment):
			return agent.ToolError{Error: "comment too long: at most 200 characters"}, nil
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
