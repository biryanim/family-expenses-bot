package usecase

import (
	"context"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

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

	ListUserIDs(ctx context.Context) ([]int64, error)
}
