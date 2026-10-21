package usecase

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (u *Usecase) UpdateExpense(ctx context.Context, id int64, p entity.ExpensePatch) (entity.Expense, error) {
	if p.Amount != nil && *p.Amount <= 0 {
		return entity.Expense{}, entity.ErrInvalidAmount
	}
	if p.Category != nil && !p.Category.Valid() {
		return entity.Expense{}, entity.ErrInvalidCategory
	}
	if p.SpentAt != nil && p.SpentAt.IsZero() {
		return entity.Expense{}, entity.ErrInvalidDate
	}
	if p.Comment != nil {
		comment := strings.TrimSpace(*p.Comment)
		if utf8.RuneCountInString(comment) > entity.MaxCommentLength {
			return entity.Expense{}, entity.ErrInvalidComment
		}
		p.Comment = &comment
	}

	updated, err := u.repo.UpdateExpense(ctx, id, p)
	if err != nil {
		return entity.Expense{}, fmt.Errorf("u.repo.UpdateExpense: %w", err)
	}
	return updated, nil
}
