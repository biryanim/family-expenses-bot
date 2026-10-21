package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (a *Agent) HandleMessage(ctx context.Context, userID int64, text string) (reply string, err error) {
	weekdays := [...]string{"воскресенье", "понедельник", "вторник", "среда", "четверг", "пятница", "суббота"}
	now := time.Now()

	definitions := make([]Tool, 0, len(a.tools))
	executors := make(map[string]ToolExecutor, len(a.tools))
	for _, executor := range a.tools {
		definition := executor.Definition()
		definitions = append(definitions, definition)
		executors[definition.Name] = executor
	}

	userMessage := Message{Role: RoleUser, Content: text}
	messages := []Message{{Role: RoleSystem, Content: fmt.Sprintf(systemPrompt, now.Format(time.DateOnly), weekdays[now.Weekday()])}}
	messages = append(messages, a.history.ListMessagesForUser(userID, now)...)
	messages = append(messages, userMessage)

	var done []string
	defer func() {
		if err == nil || len(done) == 0 {
			return
		}
		reply = strings.Join(done, " ") + " Подробнее ответить не удалось."
		a.history.AppendMessagesForUser(userID, now, userMessage, Message{Role: RoleAssistant, Content: reply})
	}()

	for range maxSteps {
		completion, err := a.llm.CreateChatCompletion(ctx, messages, definitions)
		if err != nil {
			return "", fmt.Errorf("a.llm.CreateChatCompletion: %w", err)
		}

		if len(completion.ToolCalls) == 0 {
			answer := strings.TrimSpace(completion.Content)
			if answer == "" {
				return "", entity.ErrEmptyAgentReply
			}
			a.history.AppendMessagesForUser(userID, now, userMessage, Message{Role: RoleAssistant, Content: answer})
			return answer, nil
		}

		messages = append(messages, completion)
		for _, call := range completion.ToolCalls {
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

			if verb, mutating := mutatingTools[call.Name]; mutating {
				if action, ok := result.(ActionResult); ok {
					done = append(done, action.DescribeAction()+" "+verb+".")
				}
			}
		}
	}

	return "", entity.ErrAgentStepLimit
}
