package tools_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/agent"
	"github.com/biryanim/family-expenses-bot/internal/agent/tools"
	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/internal/usecase"
	"github.com/biryanim/family-expenses-bot/mock"
)

var (
	_ tools.Usecase      = (*usecase.Usecase)(nil)
	_ agent.ToolExecutor = (*tools.AddExpense)(nil)
	_ agent.ToolExecutor = (*tools.ListExpenses)(nil)
	_ agent.ToolExecutor = (*tools.GetLastExpense)(nil)
	_ agent.ToolExecutor = (*tools.UpdateExpense)(nil)
	_ agent.ToolExecutor = (*tools.DeleteExpense)(nil)
)

const (
	userA int64 = 100
	userB int64 = 200
)

var (
	spentAt   = time.Date(2026, 10, 17, 9, 0, 0, 0, time.Local)
	createdAt = time.Date(2026, 10, 17, 9, 1, 0, 0, time.Local)
)

var userProperties = []string{"user_id", "user", "userId", "telegram_id", "member_id", "member"}

type Suite struct {
	suite.Suite

	ctrl *gomock.Controller

	uc *mock.MockToolsUsecase

	addExpense     *tools.AddExpense
	listExpenses   *tools.ListExpenses
	getLastExpense *tools.GetLastExpense
	updateExpense  *tools.UpdateExpense
	deleteExpense  *tools.DeleteExpense
}

func (s *Suite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.uc = mock.NewMockToolsUsecase(s.ctrl)
	s.addExpense = tools.NewAddExpense(s.uc)
	s.listExpenses = tools.NewListExpenses(s.uc)
	s.getLastExpense = tools.NewGetLastExpense(s.uc)
	s.updateExpense = tools.NewUpdateExpense(s.uc)
	s.deleteExpense = tools.NewDeleteExpense(s.uc)
}

func (s *Suite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *Suite) SetupSubTest() {
	s.SetupTest()
}

func (s *Suite) TearDownSubTest() {
	s.TearDownTest()
}

func TestTools(t *testing.T) {
	suite.Run(t, new(Suite))
}

func testExpense() entity.Expense {
	return entity.Expense{
		ID:        7,
		UserID:    userA,
		Amount:    35000,
		Category:  entity.CategoryFood,
		SpentAt:   spentAt,
		Comment:   "кофе",
		CreatedAt: createdAt,
	}
}

func categoryCodes() []any {
	return []any{"food", "transport", "housing", "entertainment", "health", "shopping", "other"}
}

func keys(m map[string]any) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}
