package agent_test

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

const (
	userA int64 = 100
	userB int64 = 200

	idOfUserA int64 = 7
	idOfUserB int64 = 8
	missingID int64 = 999
)

var (
	addTool  = agent.Tool{Name: "add_expense", Description: "Добавить трату", Parameters: map[string]any{"type": "object"}}
	listTool = agent.Tool{Name: "list_expenses", Description: "Показать траты", Parameters: map[string]any{"type": "object"}}
)

var (
	expenseOfUserA = entity.Expense{
		ID:        idOfUserA,
		UserID:    userA,
		Amount:    35000,
		Category:  entity.CategoryFood,
		SpentAt:   time.Date(2026, 10, 17, 9, 0, 0, 0, time.Local),
		Comment:   "кофе",
		CreatedAt: time.Date(2026, 10, 17, 9, 1, 0, 0, time.Local),
	}
	expenseOfUserB = entity.Expense{
		ID:        idOfUserB,
		UserID:    userB,
		Amount:    7777700,
		Category:  entity.CategoryEntertainment,
		SpentAt:   time.Date(2026, 9, 3, 0, 0, 0, 0, time.Local),
		Comment:   "секретный концерт",
		CreatedAt: time.Date(2026, 9, 3, 12, 0, 0, 0, time.Local),
	}
)

type Suite struct {
	suite.Suite

	ctrl *gomock.Controller

	llm   *mock.MockLLMClient
	add   *mock.MockToolExecutor
	list  *mock.MockToolExecutor
	agent *agent.Agent

	repo     *mock.MockExpenseRepository
	isolated *agent.Agent
}

func (s *Suite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.llm = mock.NewMockLLMClient(s.ctrl)
	s.add = mock.NewMockToolExecutor(s.ctrl)
	s.list = mock.NewMockToolExecutor(s.ctrl)
	s.add.EXPECT().Definition().Return(addTool).AnyTimes()
	s.list.EXPECT().Definition().Return(listTool).AnyTimes()
	s.agent = agent.New(s.llm, s.add, s.list)

	s.repo = mock.NewMockExpenseRepository(s.ctrl)
	uc := usecase.New(s.repo)
	s.isolated = agent.New(s.llm,
		tools.NewAddExpense(uc),
		tools.NewListExpenses(uc),
		tools.NewGetLastExpense(uc),
		tools.NewUpdateExpense(uc),
		tools.NewDeleteExpense(uc),
	)
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

func TestAgent(t *testing.T) {
	suite.Run(t, new(Suite))
}

func assistantText(text string) agent.Message {
	return agent.Message{Role: agent.RoleAssistant, Content: text}
}

func assistantCalls(calls ...agent.ToolCall) agent.Message {
	return agent.Message{Role: agent.RoleAssistant, ToolCalls: calls}
}
