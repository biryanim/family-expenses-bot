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

func (t *AddExpense) Execute(ctx context.Context, userID int64, arguments string) (any, error) {
	var args struct {
		Amount   json.Number `json:"amount"`
		Category string      `json:"category"`
		Comment  string      `json:"comment"`
		Date     string      `json:"date"`
	}
	decoder := json.NewDecoder(strings.NewReader(arguments))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&args); err != nil {
		return agent.ToolError{Error: "invalid arguments: " + err.Error()}, nil
	}

	amount, err := entity.ParseAmount(args.Amount.String())
	if err != nil {
		return agent.ToolError{Error: "invalid amount: expected rubles with at most two decimal places"}, nil
	}

	category := entity.Category(args.Category)
	if !category.Valid() {
		category = entity.CategoryOther
	}

	var spentAt time.Time
	if args.Date != "" {
		date, err := time.ParseInLocation(time.DateOnly, args.Date, time.Local)
		if err != nil {
			return agent.ToolError{Error: "invalid date: expected YYYY-MM-DD"}, nil
		}
		if args.Date != time.Now().Format(time.DateOnly) {
			spentAt = date
		}
	}

	created, err := t.usecase.CreateExpense(ctx, entity.Expense{
		UserID:   userID,
		Amount:   amount,
		Category: category,
		SpentAt:  spentAt,
		Comment:  strings.TrimSpace(args.Comment),
	})
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidAmount):
			return agent.ToolError{Error: "invalid amount"}, nil
		case errors.Is(err, entity.ErrInvalidCategory):
			return agent.ToolError{Error: "invalid category"}, nil
		}
		return nil, fmt.Errorf("t.usecase.CreateExpense: %w", err)
	}

	return Expense{
		ID:       created.ID,
		Amount:   entity.FormatAmount(created.Amount),
		Category: created.Category,
		Date:     created.SpentAt.In(time.Local).Format(time.DateOnly),
		Comment:  created.Comment,
	}, nil
}
