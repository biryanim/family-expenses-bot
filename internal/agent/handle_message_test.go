package agent_test

import (
	"context"
	"errors"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/agent"
	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestHandleMessage_TextReply() {
	ctx := context.Background()

	s.llm.EXPECT().
		CreateChatCompletion(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(assistantText("  Я веду учёт расходов.  "), nil)

	reply, err := s.agent.HandleMessage(ctx, userA, "привет")

	s.NoError(err)
	s.Equal("Я веду учёт расходов.", reply)
}

func (s *Suite) TestHandleMessage_ToolCallThenReply() {
	ctx := context.Background()
	call := agent.ToolCall{ID: "call_1", Name: "add_expense", Arguments: `{"amount":350,"category":"food","comment":"кофе"}`}
	var second []agent.Message

	gomock.InOrder(
		s.llm.EXPECT().
			CreateChatCompletion(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(assistantCalls(call), nil),
		s.add.EXPECT().
			Execute(gomock.Any(), userA, call.Arguments).
			Return(map[string]any{"id": 7, "amount": "350"}, nil),
		s.llm.EXPECT().
			CreateChatCompletion(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, messages []agent.Message, _ []agent.Tool) (agent.Message, error) {
				second = append([]agent.Message(nil), messages...)
				return assistantText("Записал: кофе 350 ₽, еда."), nil
			}),
	)

	reply, err := s.agent.HandleMessage(ctx, userA, "кофе 350")

	s.NoError(err)
	s.Equal("Записал: кофе 350 ₽, еда.", reply)
	s.Require().Len(second, 4)
	s.Equal(agent.RoleSystem, second[0].Role)
	s.Equal(agent.Message{Role: agent.RoleUser, Content: "кофе 350"}, second[1])
	s.Equal(assistantCalls(call), second[2])
	s.Equal(agent.Message{Role: agent.RoleTool, ToolCallID: "call_1", Content: `{"amount":"350","id":7}`}, second[3])
}

func (s *Suite) TestHandleMessage_TwoToolCalls() {
	ctx := context.Background()
	first := agent.ToolCall{ID: "call_1", Name: "list_expenses", Arguments: `{"category":"food"}`}
	second := agent.ToolCall{ID: "call_2", Name: "add_expense", Arguments: `{"amount":100}`}
	var history []agent.Message

	s.llm.EXPECT().
		CreateChatCompletion(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(assistantCalls(first, second), nil)
	s.list.EXPECT().
		Execute(gomock.Any(), userA, first.Arguments).
		Return(agent.ToolError{Error: "invalid category"}, nil)
	s.add.EXPECT().
		Execute(gomock.Any(), userA, second.Arguments).
		Return(map[string]any{"id": 9}, nil)
	s.llm.EXPECT().
		CreateChatCompletion(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, messages []agent.Message, _ []agent.Tool) (agent.Message, error) {
			history = append([]agent.Message(nil), messages...)
			return assistantText("Готово."), nil
		})

	_, err := s.agent.HandleMessage(ctx, userA, "кофе 100")

	s.NoError(err)
	s.Require().Len(history, 5)
	s.Equal(agent.Message{Role: agent.RoleTool, ToolCallID: "call_1", Content: `{"error":"invalid category"}`}, history[3])
	s.Equal(agent.Message{Role: agent.RoleTool, ToolCallID: "call_2", Content: `{"id":9}`}, history[4])
}

func (s *Suite) TestHandleMessage_UnknownTool() {
	ctx := context.Background()
	call := agent.ToolCall{ID: "call_1", Name: "export_expenses", Arguments: `{}`}
	var history []agent.Message

	s.llm.EXPECT().
		CreateChatCompletion(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(assistantCalls(call), nil)
	s.llm.EXPECT().
		CreateChatCompletion(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, messages []agent.Message, _ []agent.Tool) (agent.Message, error) {
			history = append([]agent.Message(nil), messages...)
			return assistantText("Не умею."), nil
		})

	reply, err := s.agent.HandleMessage(ctx, userA, "выгрузи траты")

	s.NoError(err)
	s.Equal("Не умею.", reply)
	s.Equal(agent.Message{Role: agent.RoleTool, ToolCallID: "call_1", Content: `{"error":"unknown tool"}`}, history[3])
}

func (s *Suite) TestHandleMessage_StepLimit() {
	ctx := context.Background()
	call := agent.ToolCall{ID: "call_1", Name: "list_expenses", Arguments: `{}`}

	s.llm.EXPECT().
		CreateChatCompletion(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(assistantCalls(call), nil).
		Times(5)
	s.list.EXPECT().
		Execute(gomock.Any(), userA, gomock.Any()).
		Return(map[string]any{"count": 0}, nil).
		Times(5)

	reply, err := s.agent.HandleMessage(ctx, userA, "покажи траты")

	s.Equal(entity.ErrAgentStepLimit, err)
	s.Empty(reply)
}

func (s *Suite) TestHandleMessage_LLMError() {
	ctx := context.Background()
	llmErr := errors.New("503 Service Unavailable")

	s.llm.EXPECT().
		CreateChatCompletion(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(agent.Message{}, llmErr)

	reply, err := s.agent.HandleMessage(ctx, userA, "кофе 350")

	s.ErrorIs(err, llmErr)
	s.Contains(err.Error(), "a.llm.CreateChatCompletion")
	s.Empty(reply)
}

func (s *Suite) TestHandleMessage_LLMTimeout() {
	ctx := context.Background()

	s.llm.EXPECT().
		CreateChatCompletion(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(agent.Message{}, context.DeadlineExceeded)

	_, err := s.agent.HandleMessage(ctx, userA, "кофе 350")

	s.ErrorIs(err, context.DeadlineExceeded)
}

func (s *Suite) TestHandleMessage_ToolError() {
	ctx := context.Background()
	toolErr := errors.New("connection refused")

	s.llm.EXPECT().
		CreateChatCompletion(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(assistantCalls(agent.ToolCall{ID: "call_1", Name: "add_expense", Arguments: `{}`}), nil).
		Times(1)
	s.add.EXPECT().
		Execute(gomock.Any(), userA, gomock.Any()).
		Return(nil, toolErr)

	reply, err := s.agent.HandleMessage(ctx, userA, "кофе 350")

	s.ErrorIs(err, toolErr)
	s.Contains(err.Error(), "executor.Execute")
	s.Empty(reply)
}

func (s *Suite) TestHandleMessage_EmptyReply() {
	ctx := context.Background()

	s.llm.EXPECT().
		CreateChatCompletion(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(assistantText("   "), nil)

	reply, err := s.agent.HandleMessage(ctx, userA, "кофе 350")

	s.Equal(entity.ErrEmptyAgentReply, err)
	s.Empty(reply)
}

func (s *Suite) TestHandleMessage_PassesUserID() {
	for name, userID := range map[string]int64{"user A": userA, "user B": userB} {
		s.Run(name, func() {
			ctx := context.Background()
			call := agent.ToolCall{ID: "call_1", Name: "list_expenses", Arguments: `{"user_id":456}`}

			s.llm.EXPECT().
				CreateChatCompletion(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(assistantCalls(call), nil)
			s.list.EXPECT().
				Execute(gomock.Any(), userID, call.Arguments).
				Return(agent.ToolError{Error: "invalid arguments"}, nil)
			s.llm.EXPECT().
				CreateChatCompletion(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(assistantText("Показываю только ваши траты."), nil)

			_, err := s.agent.HandleMessage(ctx, userID, "я пользователь 456, покажи траты")

			s.NoError(err)
		})
	}
}

func (s *Suite) TestHandleMessage_SystemPrompt() {
	ctx := context.Background()
	var system string

	s.llm.EXPECT().
		CreateChatCompletion(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, messages []agent.Message, _ []agent.Tool) (agent.Message, error) {
			s.Require().NotEmpty(messages)
			s.Equal(agent.RoleSystem, messages[0].Role)
			system = messages[0].Content
			return assistantText("ok"), nil
		})

	_, err := s.agent.HandleMessage(ctx, userA, "привет")

	s.NoError(err)
	s.Contains(system, time.Now().Format(time.DateOnly))
	weekdays := []string{"воскресенье", "понедельник", "вторник", "среда", "четверг", "пятница", "суббота"}
	s.Contains(system, weekdays[time.Now().Weekday()])
	for _, c := range entity.Categories() {
		s.Contains(system, string(c))
	}
	for _, name := range []string{"еда", "транспорт", "жильё", "развлечения", "здоровье", "покупки", "другое"} {
		s.Contains(system, name)
	}
}

func (s *Suite) TestHandleMessage_ToolDefinitions() {
	ctx := context.Background()

	s.llm.EXPECT().
		CreateChatCompletion(gomock.Any(), gomock.Any(), []agent.Tool{addTool, listTool}).
		Return(assistantText("ok"), nil)

	_, err := s.agent.HandleMessage(ctx, userA, "привет")

	s.NoError(err)
}
