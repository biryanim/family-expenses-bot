package history_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/biryanim/family-expenses-bot/internal/agent"
	"github.com/biryanim/family-expenses-bot/internal/agent/history"
)

var _ agent.ConversationStore = (*history.History)(nil)

const (
	userA int64 = 100
	userB int64 = 200
)

var now = time.Date(2026, 10, 19, 12, 0, 0, 0, time.Local)

type Suite struct {
	suite.Suite

	history *history.History
}

func (s *Suite) SetupTest() {
	s.history = history.New()
}

func TestHistory(t *testing.T) {
	suite.Run(t, new(Suite))
}

func userText(text string) agent.Message {
	return agent.Message{Role: agent.RoleUser, Content: text}
}

func assistantText(text string) agent.Message {
	return agent.Message{Role: agent.RoleAssistant, Content: text}
}
