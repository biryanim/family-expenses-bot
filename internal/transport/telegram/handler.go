package telegram

import "log/slog"

const (
	replyTextOnly = "Я понимаю только текстовые сообщения."
	replyFailure  = "Что-то пошло не так, попробуйте позже."
)

type Handler struct {
	usecase Usecase
	agent   Agent
	sender  Sender
	logger  *slog.Logger
}

func New(u Usecase, a Agent, s Sender, logger *slog.Logger) *Handler {
	return &Handler{usecase: u, agent: a, sender: s, logger: logger}
}
