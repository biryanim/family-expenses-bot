package agent_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/agent"
	"github.com/biryanim/family-expenses-bot/mock"
)

const (
	userA int64 = 100
	userB int64 = 200
)

var (
	addTool  = agent.Tool{Name: "add_expense", Description: "Добавить трату", Parameters: map[string]any{"type": "object"}}
	listTool = agent.Tool{Name: "list_expenses", Description: "Показать траты", Parameters: map[string]any{"type": "object"}}
)

type Suite struct {
	suite.Suite

	ctrl *gomock.Controller

	llm   *mock.MockLLMClient
	add   *mock.MockToolExecutor
	list  *mock.MockToolExecutor
	agent *agent.Agent
}

func (s *Suite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.llm = mock.NewMockLLMClient(s.ctrl)
	s.add = mock.NewMockToolExecutor(s.ctrl)
	s.list = mock.NewMockToolExecutor(s.ctrl)
	s.add.EXPECT().Definition().Return(addTool).AnyTimes()
	s.list.EXPECT().Definition().Return(listTool).AnyTimes()
	s.agent = agent.New(s.llm, s.add, s.list)
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
