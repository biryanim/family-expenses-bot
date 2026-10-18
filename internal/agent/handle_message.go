package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (a *Agent) HandleMessage(ctx context.Context, userID int64, text string) (string, error) {
	weekdays := [...]string{"воскресенье", "понедельник", "вторник", "среда", "четверг", "пятница", "суббота"}
	now := time.Now()

	definitions := make([]Tool, 0, len(a.tools))
	executors := make(map[string]ToolExecutor, len(a.tools))
	for _, executor := range a.tools {
		definition := executor.Definition()
		definitions = append(definitions, definition)
		executors[definition.Name] = executor
	}

	messages := []Message{
		{Role: RoleSystem, Content: fmt.Sprintf(systemPrompt, now.Format(time.DateOnly), weekdays[now.Weekday()])},
		{Role: RoleUser, Content: text},
	}

	for range maxSteps {
		reply, err := a.llm.CreateChatCompletion(ctx, messages, definitions)
		if err != nil {
			return "", fmt.Errorf("a.llm.CreateChatCompletion: %w", err)
		}

		if len(reply.ToolCalls) == 0 {
			answer := strings.TrimSpace(reply.Content)
			if answer == "" {
				return "", entity.ErrEmptyAgentReply
			}
			return answer, nil
		}

		messages = append(messages, reply)
		for _, call := range reply.ToolCalls {
			var result any = ToolError{Error: "unknown tool"}
			if executor, ok := executors[call.Name]; ok {
				result, err = executor.Execute(ctx, userID, call.Arguments)
				if err != nil {
					return "", fmt.Errorf("executor.Execute: %w", err)
				}
			}

			content, err := json.Marshal(result)
			if err != nil {
				return "", fmt.Errorf("json.Marshal: %w", err)
			}
			messages = append(messages, Message{Role: RoleTool, Content: string(content), ToolCallID: call.ID})
		}
	}

	return "", entity.ErrAgentStepLimit
}
