package usecase_test

import (
	"context"
	"errors"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestRegisterMember_Success() {
	ctx := context.Background()
	user := entity.User{TelegramID: userA, Username: "ivan", FirstName: "Иван", LastName: "Петров"}

	s.repo.EXPECT().
		UpsertUser(gomock.Any(), user).
		Return(nil)

	err := s.uc.RegisterMember(ctx, user)

	s.NoError(err)
}

func (s *Suite) TestRegisterMember_RepositoryError() {
	ctx := context.Background()
	repoErr := errors.New("connection refused")

	s.repo.EXPECT().
		UpsertUser(gomock.Any(), gomock.Any()).
		Return(repoErr)

	err := s.uc.RegisterMember(ctx, entity.User{TelegramID: userA})

	s.Error(err)
	s.ErrorIs(err, repoErr)
	s.Contains(err.Error(), "u.repo.UpsertUser")
}
