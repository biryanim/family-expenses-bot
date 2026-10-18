package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"

	"github.com/biryanim/family-expenses-bot/internal/agent"
	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (c *Client) CreateChatCompletion(ctx context.Context, messages []agent.Message, tools []agent.Tool) (agent.Message, error) {
	params := openai.ChatCompletionNewParams{
		Model:    c.model,
		Messages: make([]openai.ChatCompletionMessageParamUnion, 0, len(messages)),
	}

	for _, m := range messages {
		switch m.Role {
		case agent.RoleSystem:
			params.Messages = append(params.Messages, openai.SystemMessage(m.Content))
		case agent.RoleUser:
			params.Messages = append(params.Messages, openai.UserMessage(m.Content))
		case agent.RoleTool:
			params.Messages = append(params.Messages, openai.ToolMessage(m.Content, m.ToolCallID))
		case agent.RoleAssistant:
			assistant := openai.ChatCompletionAssistantMessageParam{}
			if m.Content != "" {
				assistant.Content.OfString = openai.String(m.Content)
			}
			for _, tc := range m.ToolCalls {
				assistant.ToolCalls = append(assistant.ToolCalls, openai.ChatCompletionMessageToolCallUnionParam{
					OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
						ID: tc.ID,
						Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
							Name:      tc.Name,
							Arguments: tc.Arguments,
						},
					},
				})
			}
			params.Messages = append(params.Messages, openai.ChatCompletionMessageParamUnion{OfAssistant: &assistant})
		default:
			return agent.Message{}, fmt.Errorf("unknown message role %q", m.Role)
		}
	}

	for _, t := range tools {
		params.Tools = append(params.Tools, openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        t.Name,
			Description: openai.String(t.Description),
			Parameters:  shared.FunctionParameters(t.Parameters),
		}))
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	completion, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return agent.Message{}, fmt.Errorf("c.client.Chat.Completions.New: %w", err)
	}
	if len(completion.Choices) == 0 {
		return agent.Message{}, entity.ErrEmptyCompletion
	}

	choice := completion.Choices[0].Message
	result := agent.Message{
		Role:    agent.RoleAssistant,
		Content: choice.Content,
	}
	for _, tc := range choice.ToolCalls {
		if tc.Type != "function" {
			continue
		}
		args := tc.Function.Arguments
		if strings.TrimSpace(args) == "" {
			args = "{}"
		}
		result.ToolCalls = append(result.ToolCalls, agent.ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: args,
		})
	}
	return result, nil
}
