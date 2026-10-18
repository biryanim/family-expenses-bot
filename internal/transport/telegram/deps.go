package telegram

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

//go:generate go tool mockgen -source=./deps.go -destination ../../../mock/telegram.go -package mock -mock_names Usecase=MockTelegramUsecase
type Usecase interface {
	RegisterMember(ctx context.Context, u entity.User) error
}

type Agent interface {
	HandleMessage(ctx context.Context, userID int64, text string) (string, error)
}

type Sender interface {
	SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error)
}
