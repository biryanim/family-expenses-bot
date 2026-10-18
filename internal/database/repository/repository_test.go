package repository_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/database/repository"
	"github.com/biryanim/family-expenses-bot/internal/entity"
	"github.com/biryanim/family-expenses-bot/internal/usecase"
	"github.com/biryanim/family-expenses-bot/mock"
)

var _ usecase.ExpenseRepository = (*repository.Repository)(nil)

const (
	userA int64 = 100
	userB int64 = 200

	selectExpenses = "SELECT id, user_id, amount, category, spent_at, comment, created_at, updated_at FROM expenses"
	orderExpenses  = " ORDER BY spent_at DESC, id DESC"
)

var (
	spentAt   = time.Date(2026, 10, 17, 12, 30, 0, 0, time.FixedZone("MSK", 3*60*60))
	createdAt = time.Date(2026, 10, 17, 12, 31, 0, 0, time.UTC)
	updatedAt = time.Date(2026, 10, 18, 9, 0, 0, 0, time.UTC)
	from      = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to        = time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
)

type Suite struct {
	suite.Suite

	ctrl *gomock.Controller

	pool *mock.MockPool
	repo *repository.Repository
}

func (s *Suite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.pool = mock.NewMockPool(s.ctrl)
	s.repo = repository.New(s.pool)
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

func TestRepository(t *testing.T) {
	suite.Run(t, new(Suite))
}

func mockAnyArgs(count int) []any {
	args := make([]any, count)

	for i := range args {
		args[i] = gomock.Any()
	}

	return args
}

func expenseScanArgs() []any {
	return mockAnyArgs(8)
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

func populateExpenseScanDest(e entity.Expense, dest ...any) {
	*dest[0].(*int64) = e.ID
	*dest[1].(*int64) = e.UserID
	*dest[2].(*int64) = e.Amount
	*dest[3].(*entity.Category) = e.Category
	*dest[4].(*time.Time) = e.SpentAt
	*dest[5].(*string) = e.Comment
	*dest[6].(*time.Time) = e.CreatedAt
	*dest[7].(**time.Time) = e.UpdatedAt
}
