package usecase_test

import (
	"context"
	"errors"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestListMembers_Success() {
	ctx := context.Background()
	users := []entity.User{
		{TelegramID: 50, FirstName: "Анна", CreatedAt: createdAt},
		{TelegramID: 100, Username: "petr", FirstName: "Пётр", LastName: "Иванов", CreatedAt: createdAt},
	}

	s.repo.EXPECT().
		ListUsers(gomock.Any()).
		Return(users, nil)

	result, err := s.uc.ListMembers(ctx)

	s.NoError(err)
	s.Equal(users, result)
}

func (s *Suite) TestListMembers_Empty() {
	ctx := context.Background()

	s.repo.EXPECT().
		ListUsers(gomock.Any()).
		Return([]entity.User{}, nil)

	result, err := s.uc.ListMembers(ctx)

	s.NoError(err)
	s.NotNil(result)
	s.Empty(result)
}

func (s *Suite) TestListMembers_RepositoryError() {
	ctx := context.Background()
	repoErr := errors.New("connection refused")

	s.repo.EXPECT().
		ListUsers(gomock.Any()).
		Return(nil, repoErr)

	result, err := s.uc.ListMembers(ctx)

	s.Error(err)
	s.Nil(result)
	s.ErrorIs(err, repoErr)
	s.Contains(err.Error(), "u.repo.ListUsers")
}
