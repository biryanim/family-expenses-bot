package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

//go:generate go tool mockgen -source=./deps.go -destination ../../../mock/repository.go -package mock
//go:generate go tool mockgen -destination ../../../mock/row.go -package mock github.com/jackc/pgx/v5 Row
//go:generate go tool mockgen -destination ../../../mock/rows.go -package mock github.com/jackc/pgx/v5 Rows
type Pool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
