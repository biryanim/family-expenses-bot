package agent

import "context"

//go:generate go tool mockgen -source=./deps.go -destination ../../mock/agent.go -package mock
type LLMClient interface {
	CreateChatCompletion(ctx context.Context, messages []Message, tools []Tool) (Message, error)
}
