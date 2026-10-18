package telegram

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (h *Handler) HandleUpdate(ctx context.Context, _ *bot.Bot, update *models.Update) {
	var userID int64
	defer func() {
		if r := recover(); r != nil {
			h.logger.ErrorContext(ctx, "telegram update panicked", slog.Int64("user_id", userID), slog.Any("panic", r))
		}
	}()

	if update.Message == nil || update.Message.From == nil {
		return
	}
	msg := update.Message
	userID = msg.From.ID
	chatID := msg.Chat.ID

	if msg.Text == "" {
		if _, err := h.sender.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: replyTextOnly}); err != nil {
			h.logger.ErrorContext(ctx, "telegram send message failed",
				slog.Int64("user_id", userID), slog.Int64("chat_id", chatID), slog.Any("error", err))
		}
		return
	}

	member := entity.User{
		TelegramID: userID,
		Username:   msg.From.Username,
		FirstName:  msg.From.FirstName,
		LastName:   msg.From.LastName,
	}
	if err := h.usecase.RegisterMember(ctx, member); err != nil {
		h.logger.ErrorContext(ctx, "register member failed", slog.Int64("user_id", userID), slog.Any("error", err))
		if _, err := h.sender.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: replyFailure}); err != nil {
			h.logger.ErrorContext(ctx, "telegram send message failed",
				slog.Int64("user_id", userID), slog.Int64("chat_id", chatID), slog.Any("error", err))
		}
		return
	}

	reply, err := h.agent.HandleMessage(ctx, userID, msg.Text)
	if err != nil {
		h.logger.ErrorContext(ctx, "agent handle message failed", slog.Int64("user_id", userID), slog.Any("error", err))
		reply = replyFailure
	}

	if _, err := h.sender.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: reply}); err != nil {
		h.logger.ErrorContext(ctx, "telegram send message failed",
			slog.Int64("user_id", userID), slog.Int64("chat_id", chatID), slog.Any("error", err))
	}
}
