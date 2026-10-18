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

func (t *GetLastExpense) Execute(ctx context.Context, userID int64, arguments string) (any, error) {
	var args struct{}
	decoder := json.NewDecoder(strings.NewReader(arguments))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&args); err != nil {
		return agent.ToolError{Error: "invalid arguments: " + err.Error()}, nil
	}

	last, err := t.usecase.GetLastExpenseForUser(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return agent.ToolError{Error: "expense not found"}, nil
		}
		return nil, fmt.Errorf("t.usecase.GetLastExpenseForUser: %w", err)
	}

	return Expense{
		ID:       last.ID,
		Amount:   entity.FormatAmount(last.Amount),
		Category: last.Category,
		Date:     last.SpentAt.In(time.Local).Format(time.DateOnly),
		Comment:  last.Comment,
	}, nil
}
