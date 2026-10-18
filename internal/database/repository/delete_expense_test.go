package repository_test

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestDeleteExpense_Success() {
	ctx := context.Background()
	id := int64(7)

	s.pool.EXPECT().
		Exec(gomock.Any(), "DELETE FROM expenses WHERE id = $1", id).
		Return(pgconn.NewCommandTag("DELETE 1"), nil)

	err := s.repo.DeleteExpense(ctx, id)

	s.NoError(err)
}

func (s *Suite) TestDeleteExpense_1OTFOUND() {
	ctx := context.Background()
	id := int64(999)

	s.pool.EXPECT().
		Exec(gomock.Any(), "DELETE FROM expenses WHERE id = $1", id).
		Return(pgconn.NewCommandTag("DELETE 0"), nil)

	err := s.repo.DeleteExpense(ctx, id)

	s.Error(err)
	s.Equal(entity.ErrNotFound, err)
}

func (s *Suite) TestDeleteExpense_ExecError() {
	ctx := context.Background()
	id := int64(7)
	execErr := errors.New("exec error")

	s.pool.EXPECT().
		Exec(gomock.Any(), gomock.Any(), mockAnyArgs(1)...).
		Return(pgconn.CommandTag{}, execErr)

	err := s.repo.DeleteExpense(ctx, id)

	s.Error(err)
	s.ErrorIs(err, execErr)
	s.Contains(err.Error(), "r.pool.Exec")
}
