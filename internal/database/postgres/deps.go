package postgres

type Logger interface {
	Info(msg string, args ...any)
}
