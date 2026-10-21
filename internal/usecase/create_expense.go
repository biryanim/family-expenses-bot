package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (u *Usecase) CreateExpense(ctx context.Context, e entity.Expense) (entity.Expense, error) {
	if e.Amount <= 0 {
		return entity.Expense{}, entity.ErrInvalidAmount
	}
	if !e.Category.Valid() {
		return entity.Expense{}, entity.ErrInvalidCategory
	}
	e.Comment = strings.TrimSpace(e.Comment)
	if utf8.RuneCountInString(e.Comment) > entity.MaxCommentLength {
		return entity.Expense{}, entity.ErrInvalidComment
	}
	if e.SpentAt.IsZero() {
		e.SpentAt = time.Now()
	}

	created, err := u.repo.CreateExpense(ctx, e)
	if err != nil {
		return entity.Expense{}, fmt.Errorf("u.repo.CreateExpense: %w", err)
	}
	return created, nil
}
