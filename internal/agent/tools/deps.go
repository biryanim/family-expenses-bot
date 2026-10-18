package tools

import (
	"context"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

//go:generate go tool mockgen -source=./deps.go -destination ../../../mock/tools.go -package mock -mock_names Usecase=MockToolsUsecase
type Usecase interface {
	CreateExpense(ctx context.Context, e entity.Expense) (entity.Expense, error)
	ListExpensesForUser(ctx context.Context, userID int64, f entity.ExpenseFilter) ([]entity.Expense, error)
	GetExpenseForUser(ctx context.Context, userID, id int64) (entity.Expense, error)
	GetLastExpenseForUser(ctx context.Context, userID int64) (entity.Expense, error)
	UpdateExpenseForUser(ctx context.Context, userID int64, e entity.Expense) (entity.Expense, error)
	DeleteExpenseForUser(ctx context.Context, userID, id int64) error
}
