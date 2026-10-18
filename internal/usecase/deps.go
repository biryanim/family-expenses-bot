package usecase

import (
	"context"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

//go:generate go tool mockgen -source=./deps.go -destination ../../mock/usecase.go -package mock
type ExpenseRepository interface {
	CreateExpense(ctx context.Context, e entity.Expense) (entity.Expense, error)
	GetExpense(ctx context.Context, id int64) (entity.Expense, error)
	ListExpenses(ctx context.Context, f entity.ExpenseFilter) ([]entity.Expense, error)
	UpdateExpense(ctx context.Context, e entity.Expense) (entity.Expense, error)
	DeleteExpense(ctx context.Context, id int64) error

	GetExpenseForUser(ctx context.Context, userID, id int64) (entity.Expense, error)
	ListExpensesForUser(ctx context.Context, userID int64, f entity.ExpenseFilter) ([]entity.Expense, error)
	UpdateExpenseForUser(ctx context.Context, userID int64, e entity.Expense) (entity.Expense, error)
	DeleteExpenseForUser(ctx context.Context, userID, id int64) error
	GetLastExpenseForUser(ctx context.Context, userID int64) (entity.Expense, error)

	ListUsers(ctx context.Context) ([]entity.User, error)
	UpsertUser(ctx context.Context, u entity.User) error
}
