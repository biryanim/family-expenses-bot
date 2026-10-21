package tools_test

import (
	"github.com/biryanim/family-expenses-bot/internal/agent"
	"github.com/biryanim/family-expenses-bot/internal/agent/tools"
)

var _ agent.ActionResult = tools.Expense{}

func (s *Suite) TestDescribeAction_Expense() {
	expense := tools.Expense{ID: 42, Amount: "350.50", Date: "2026-10-21"}

	s.Equal("Трата на 350.50 ₽ за 2026-10-21", expense.DescribeAction())
}
