package agent

import "context"

//go:generate go tool mockgen -source=./deps.go -destination ../../mock/agent.go -package mock
type LLMClient interface {
	CreateChatCompletion(ctx context.Context, messages []Message, tools []Tool) (Message, error)
}

type ToolExecutor interface {
	Definition() Tool
	Execute(ctx context.Context, userID int64, arguments string) (any, error)
}
