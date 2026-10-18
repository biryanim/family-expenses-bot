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

func (t *DeleteExpense) Execute(ctx context.Context, userID int64, arguments string) (any, error) {
	var args struct {
		ID int64 `json:"id"`
	}
	decoder := json.NewDecoder(strings.NewReader(arguments))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&args); err != nil {
		return agent.ToolError{Error: "invalid arguments: " + err.Error()}, nil
	}

	expense, err := t.usecase.GetExpenseForUser(ctx, userID, args.ID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return agent.ToolError{Error: "expense not found"}, nil
		}
		return nil, fmt.Errorf("t.usecase.GetExpenseForUser: %w", err)
	}

	if err := t.usecase.DeleteExpenseForUser(ctx, userID, expense.ID); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return agent.ToolError{Error: "expense not found"}, nil
		}
		return nil, fmt.Errorf("t.usecase.DeleteExpenseForUser: %w", err)
	}

	return Expense{
		ID:       expense.ID,
		Amount:   entity.FormatAmount(expense.Amount),
		Category: expense.Category,
		Date:     expense.SpentAt.In(time.Local).Format(time.DateOnly),
		Comment:  expense.Comment,
	}, nil
}
