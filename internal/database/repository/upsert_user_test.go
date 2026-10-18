package repository_test

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

const upsertUserSQL = "INSERT INTO users (telegram_id, username, first_name, last_name) VALUES ($1, $2, $3, $4)" +
	" ON CONFLICT (telegram_id) DO UPDATE SET username = EXCLUDED.username, first_name = EXCLUDED.first_name," +
	" last_name = EXCLUDED.last_name, updated_at = now()" +
	" WHERE (users.username, users.first_name, users.last_name)" +
	" IS DISTINCT FROM (EXCLUDED.username, EXCLUDED.first_name, EXCLUDED.last_name)"

func (s *Suite) TestUpsertUser_Success() {
	ctx := context.Background()
	user := entity.User{TelegramID: 123, Username: "ivan", FirstName: "Иван"}

	s.pool.EXPECT().
		Exec(gomock.Any(), upsertUserSQL, int64(123), "ivan", "Иван", "").
		Return(pgconn.NewCommandTag("INSERT 0 1"), nil)

	err := s.repo.UpsertUser(ctx, user)

	s.NoError(err)
}

func (s *Suite) TestUpsertUser_Unchanged() {
	ctx := context.Background()
	user := entity.User{TelegramID: 123, Username: "ivan", FirstName: "Иван"}

	s.pool.EXPECT().
		Exec(gomock.Any(), upsertUserSQL, int64(123), "ivan", "Иван", "").
		Return(pgconn.NewCommandTag("INSERT 0 0"), nil)

	err := s.repo.UpsertUser(ctx, user)

	s.NoError(err)
}

func (s *Suite) TestUpsertUser_ExecError() {
	ctx := context.Background()
	execErr := errors.New("connection refused")

	s.pool.EXPECT().
		Exec(gomock.Any(), gomock.Any(), mockAnyArgs(4)...).
		Return(pgconn.CommandTag{}, execErr)

	err := s.repo.UpsertUser(ctx, entity.User{TelegramID: 123})

	s.Error(err)
	s.ErrorIs(err, execErr)
	s.Contains(err.Error(), "r.pool.Exec")
}
