package repository_test

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestDeleteExpenseForUser_Success() {
	ctx := context.Background()
	id := int64(7)

	s.pool.EXPECT().
		Exec(gomock.Any(), "DELETE FROM expenses WHERE id = $1 AND user_id = $2", id, userA).
		Return(pgconn.NewCommandTag("DELETE 1"), nil)

	err := s.repo.DeleteExpenseForUser(ctx, userA, id)

	s.NoError(err)
}

func (s *Suite) TestDeleteExpenseForUser_2OTFOUND() {
	ctx := context.Background()
	id := int64(8)

	s.pool.EXPECT().
		Exec(gomock.Any(), "DELETE FROM expenses WHERE id = $1 AND user_id = $2", id, userA).
		Return(pgconn.NewCommandTag("DELETE 0"), nil)

	err := s.repo.DeleteExpenseForUser(ctx, userA, id)

	s.Error(err)
	s.Equal(entity.ErrNotFound, err)
}

func (s *Suite) TestDeleteExpenseForUser_ExecError() {
	ctx := context.Background()
	id := int64(7)
	execErr := errors.New("exec error")

	s.pool.EXPECT().
		Exec(gomock.Any(), gomock.Any(), mockAnyArgs(2)...).
		Return(pgconn.CommandTag{}, execErr)

	err := s.repo.DeleteExpenseForUser(ctx, userA, id)

	s.Error(err)
	s.ErrorIs(err, execErr)
	s.Contains(err.Error(), "r.pool.Exec")
}
