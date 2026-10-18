package telegram

import "log/slog"

const (
	replyTextOnly = "Я понимаю только текстовые сообщения."
	replyFailure  = "Что-то пошло не так, попробуйте позже."
)

type Handler struct {
	usecase Usecase
	sender  Sender
	logger  *slog.Logger
}

func New(u Usecase, s Sender, logger *slog.Logger) *Handler {
	return &Handler{usecase: u, sender: s, logger: logger}
}
