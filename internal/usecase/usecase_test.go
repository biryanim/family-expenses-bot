package usecase_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/internal/usecase"
	"github.com/biryanim/family-expenses-bot/mock"
)

const userA int64 = 100

var (
	spentAt   = time.Date(2026, 10, 17, 9, 0, 0, 0, time.FixedZone("MSK", 3*60*60))
	updatedAt = time.Date(2026, 10, 18, 10, 0, 0, 0, time.FixedZone("MSK", 3*60*60))
	createdAt = time.Date(2026, 10, 17, 9, 1, 0, 0, time.UTC)
	from      = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to        = time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
)

type Suite struct {
	suite.Suite

	ctrl *gomock.Controller

	repo *mock.MockExpenseRepository
	uc   *usecase.Usecase
}

func (s *Suite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.repo = mock.NewMockExpenseRepository(s.ctrl)
	s.uc = usecase.New(s.repo)
}

func (s *Suite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *Suite) SetupSubTest() {
	s.SetupTest()
}

func (s *Suite) TearDownSubTest() {
	s.TearDownTest()
}

func TestUsecase(t *testing.T) {
	suite.Run(t, new(Suite))
}

func testExpense() entity.Expense {
	return entity.Expense{
		ID:        7,
		UserID:    userA,
		Amount:    35000,
		Category:  entity.CategoryFood,
		SpentAt:   spentAt,
		Comment:   "кофе",
		CreatedAt: createdAt,
	}
}
