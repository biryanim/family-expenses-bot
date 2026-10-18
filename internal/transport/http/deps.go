package http

import (
	"context"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

//go:generate go tool mockgen -source=./deps.go -destination ../../../mock/http.go -package mock
type Usecase interface {
	CreateExpense(ctx context.Context, e entity.Expense) (entity.Expense, error)
	GetExpense(ctx context.Context, id int64) (entity.Expense, error)
	ListExpenses(ctx context.Context, f entity.ExpenseFilter) ([]entity.Expense, error)
	UpdateExpense(ctx context.Context, e entity.Expense) (entity.Expense, error)
	DeleteExpense(ctx context.Context, id int64) error
	ListMembers(ctx context.Context) ([]entity.User, error)
}
