package telegram

import "log/slog"

const (
	replyTextOnly = "Я понимаю только текстовые сообщения."
	replyFailure  = "Что-то пошло не так, попробуйте позже."
	replyTooLong  = "Сообщение слишком длинное, сократите его до 1000 символов."
)

type Handler struct {
	usecase Usecase
	agent   Agent
	sender  Sender
	logger  *slog.Logger
	locks   *UserLocks
}

func New(u Usecase, a Agent, s Sender, logger *slog.Logger) *Handler {
	return &Handler{usecase: u, agent: a, sender: s, logger: logger, locks: NewUserLocks()}
}
