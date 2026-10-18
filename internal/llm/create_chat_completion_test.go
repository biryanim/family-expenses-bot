package llm_test

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/openai/openai-go/v3"

	"github.com/biryanim/family-expenses-bot/internal/agent"
	"github.com/biryanim/family-expenses-bot/internal/entity"
)

var addExpenseTool = agent.Tool{
	Name:        "add_expense",
	Description: "Добавить трату",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"amount": map[string]any{"type": "number"},
		},
		"required": []any{"amount"},
	},
}

var userMessages = []agent.Message{{Role: agent.RoleUser, Content: "кофе 350"}}

func (s *Suite) TestCreateChatCompletion_Request() {
	ctx := context.Background()
	s.handle = respondJSON(http.StatusOK, completionJSON(`{"role":"assistant","content":"ok"}`))

	_, err := s.client.CreateChatCompletion(ctx, userMessages, []agent.Tool{addExpenseTool})

	s.Require().NoError(err)
	reqs := s.recorded()
	s.Require().Len(reqs, 1)
	s.Equal("/v1/chat/completions", reqs[0].path)
	s.Equal("Bearer "+apiKey, reqs[0].auth)
	s.Equal(model, reqs[0].body["model"])
	tools, ok := reqs[0].body["tools"].([]any)
	s.Require().True(ok)
	s.Require().Len(tools, 1)
	tool := tools[0].(map[string]any)
	s.Equal("function", tool["type"])
	function := tool["function"].(map[string]any)
	s.Equal("add_expense", function["name"])
	s.Equal("Добавить трату", function["description"])
	s.Equal(map[string]any{
		"type":       "object",
		"properties": map[string]any{"amount": map[string]any{"type": "number"}},
		"required":   []any{"amount"},
	}, function["parameters"])
}

func (s *Suite) TestCreateChatCompletion_NoTools() {
	ctx := context.Background()
	s.handle = respondJSON(http.StatusOK, completionJSON(`{"role":"assistant","content":"ok"}`))

	_, err := s.client.CreateChatCompletion(ctx, userMessages, nil)

	s.Require().NoError(err)
	reqs := s.recorded()
	s.Require().Len(reqs, 1)
	s.NotContains(reqs[0].body, "tools")
}

func (s *Suite) TestCreateChatCompletion_History() {
	ctx := context.Background()
	s.handle = respondJSON(http.StatusOK, completionJSON(`{"role":"assistant","content":"Добавил"}`))
	history := []agent.Message{
		{Role: agent.RoleSystem, Content: "Ты бот учёта расходов"},
		{Role: agent.RoleUser, Content: "кофе 350"},
		{Role: agent.RoleAssistant, ToolCalls: []agent.ToolCall{{ID: "call_1", Name: "add_expense", Arguments: `{"amount":350}`}}},
		{Role: agent.RoleTool, ToolCallID: "call_1", Content: `{"id":1}`},
	}

	_, err := s.client.CreateChatCompletion(ctx, history, []agent.Tool{addExpenseTool})

	s.Require().NoError(err)
	reqs := s.recorded()
	s.Require().Len(reqs, 1)
	messages := reqs[0].body["messages"].([]any)
	s.Require().Len(messages, 4)
	s.Equal("system", messages[0].(map[string]any)["role"])
	s.Equal("user", messages[1].(map[string]any)["role"])
	assistant := messages[2].(map[string]any)
	s.Equal("assistant", assistant["role"])
	s.NotContains(assistant, "content")
	calls := assistant["tool_calls"].([]any)
	s.Require().Len(calls, 1)
	call := calls[0].(map[string]any)
	s.Equal("call_1", call["id"])
	s.Equal("function", call["type"])
	s.Equal(map[string]any{"name": "add_expense", "arguments": `{"amount":350}`}, call["function"])
	toolMsg := messages[3].(map[string]any)
	s.Equal("tool", toolMsg["role"])
	s.Equal("call_1", toolMsg["tool_call_id"])
	s.Equal(`{"id":1}`, toolMsg["content"])
}

func (s *Suite) TestCreateChatCompletion_ToolCall() {
	ctx := context.Background()
	s.handle = respondJSON(http.StatusOK, completionJSON(`{"role":"assistant","content":null,"tool_calls":[{"id":"call_1","type":"function","function":{"name":"add_expense","arguments":"{\"amount\":350}"}}]}`))

	msg, err := s.client.CreateChatCompletion(ctx, userMessages, []agent.Tool{addExpenseTool})

	s.Require().NoError(err)
	s.Equal(agent.RoleAssistant, msg.Role)
	s.Empty(msg.Content)
	s.Equal([]agent.ToolCall{{ID: "call_1", Name: "add_expense", Arguments: `{"amount":350}`}}, msg.ToolCalls)
}

func (s *Suite) TestCreateChatCompletion_Text() {
	ctx := context.Background()
	s.handle = respondJSON(http.StatusOK, completionJSON(`{"role":"assistant","content":"Привет!"}`))

	msg, err := s.client.CreateChatCompletion(ctx, userMessages, []agent.Tool{addExpenseTool})

	s.Require().NoError(err)
	s.Equal(agent.RoleAssistant, msg.Role)
	s.Equal("Привет!", msg.Content)
	s.Empty(msg.ToolCalls)
}

func (s *Suite) TestCreateChatCompletion_EmptyArguments() {
	ctx := context.Background()
	s.handle = respondJSON(http.StatusOK, completionJSON(`{"role":"assistant","content":null,"tool_calls":[{"id":"call_1","type":"function","function":{"name":"list_expenses","arguments":""}}]}`))

	msg, err := s.client.CreateChatCompletion(ctx, userMessages, []agent.Tool{addExpenseTool})

	s.Require().NoError(err)
	s.Equal([]agent.ToolCall{{ID: "call_1", Name: "list_expenses", Arguments: "{}"}}, msg.ToolCalls)
}

func (s *Suite) TestCreateChatCompletion_EmptyChoices() {
	ctx := context.Background()
	s.handle = respondJSON(http.StatusOK, `{"id":"chatcmpl-1","object":"chat.completion","created":1,"model":"test-model","choices":[]}`)

	msg, err := s.client.CreateChatCompletion(ctx, userMessages, nil)

	s.Equal(entity.ErrEmptyCompletion, err)
	s.Zero(msg)
}

func (s *Suite) TestCreateChatCompletion_UnknownRole() {
	ctx := context.Background()
	s.handle = respondJSON(http.StatusOK, completionJSON(`{"role":"assistant","content":"ok"}`))

	msg, err := s.client.CreateChatCompletion(ctx, []agent.Message{{Role: "developer", Content: "x"}}, nil)

	s.Require().Error(err)
	s.Contains(err.Error(), `"developer"`)
	s.Zero(msg)
	s.Empty(s.recorded())
}

func (s *Suite) TestCreateChatCompletion_RetryThenSuccess() {
	ctx := context.Background()
	s.handle = func(w http.ResponseWriter, r *http.Request) {
		if len(s.recorded()) < 3 {
			respondJSON(http.StatusServiceUnavailable, `{"error":{"message":"overloaded"}}`)(w, r)
			return
		}
		respondJSON(http.StatusOK, completionJSON(`{"role":"assistant","content":"ok"}`))(w, r)
	}

	msg, err := s.client.CreateChatCompletion(ctx, userMessages, nil)

	s.Require().NoError(err)
	s.Equal("ok", msg.Content)
	s.Len(s.recorded(), 3)
}

func (s *Suite) TestCreateChatCompletion_RetryExhausted() {
	ctx := context.Background()
	s.handle = respondJSON(http.StatusServiceUnavailable, `{"error":{"message":"overloaded"}}`)

	msg, err := s.client.CreateChatCompletion(ctx, userMessages, nil)

	s.Require().Error(err)
	s.Zero(msg)
	s.Len(s.recorded(), 3)
	apiErr, ok := errors.AsType[*openai.Error](err)
	s.Require().True(ok)
	s.Equal(http.StatusServiceUnavailable, apiErr.StatusCode)
	s.Contains(err.Error(), "c.client.Chat.Completions.New")
}

func (s *Suite) TestCreateChatCompletion_Unauthorized() {
	ctx := context.Background()
	s.handle = respondJSON(http.StatusUnauthorized, `{"error":{"message":"Incorrect API key provided: `+apiKey+`","type":"invalid_request_error","code":"invalid_api_key"}}`)

	msg, err := s.client.CreateChatCompletion(ctx, userMessages, nil)

	s.Require().Error(err)
	s.Zero(msg)
	s.Len(s.recorded(), 1)
	s.Contains(err.Error(), "401")
	s.NotContains(err.Error(), apiKey)
}

func (s *Suite) TestCreateChatCompletion_InvalidJSON() {
	ctx := context.Background()
	s.handle = respondJSON(http.StatusOK, `not json`)

	msg, err := s.client.CreateChatCompletion(ctx, userMessages, nil)

	s.Require().Error(err)
	s.Zero(msg)
	s.Contains(err.Error(), "c.client.Chat.Completions.New")
}

func (s *Suite) TestCreateChatCompletion_Timeout() {
	ctx := context.Background()
	client := s.newClient(100 * time.Millisecond)
	s.handle = func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}

	start := time.Now()
	msg, err := client.CreateChatCompletion(ctx, userMessages, nil)
	elapsed := time.Since(start)

	s.ErrorIs(err, context.DeadlineExceeded)
	s.Zero(msg)
	s.Less(elapsed, time.Second)
}

func (s *Suite) TestCreateChatCompletion_Canceled() {
	ctx, cancel := context.WithCancel(context.Background())
	received := make(chan struct{})
	s.handle = func(_ http.ResponseWriter, r *http.Request) {
		close(received)
		<-r.Context().Done()
	}
	go func() {
		<-received
		cancel()
	}()

	msg, err := s.client.CreateChatCompletion(ctx, userMessages, nil)

	s.ErrorIs(err, context.Canceled)
	s.Zero(msg)
}
