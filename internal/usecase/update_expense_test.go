package usecase_test

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestUpdateExpense_Success() {
	ctx := context.Background()
	amount := int64(150000)
	category := entity.CategoryTransport
	comment := ""
	patch := entity.ExpensePatch{Amount: &amount, Category: &category, SpentAt: &spentAt, Comment: &comment}
	updated := testExpense()
	updated.Amount = 150000
	updated.Category = entity.CategoryTransport
	updated.Comment = ""
	updated.UpdatedAt = &updatedAt

	s.repo.EXPECT().
		UpdateExpense(gomock.Any(), int64(7), patch).
		Return(updated, nil)

	result, err := s.uc.UpdateExpense(ctx, 7, patch)

	s.NoError(err)
	s.Equal(updated, result)
}

func (s *Suite) TestUpdateExpense_EmptyPatch() {
	ctx := context.Background()

	s.repo.EXPECT().
		UpdateExpense(gomock.Any(), int64(7), entity.ExpensePatch{}).
		Return(testExpense(), nil)

	result, err := s.uc.UpdateExpense(ctx, 7, entity.ExpensePatch{})

	s.NoError(err)
	s.Equal(testExpense(), result)
}

func (s *Suite) TestUpdateExpense_InvalidAmount() {
	ctx := context.Background()
	amount := int64(0)

	result, err := s.uc.UpdateExpense(ctx, 7, entity.ExpensePatch{Amount: &amount})

	s.Equal(entity.ErrInvalidAmount, err)
	s.Zero(result)
}

func (s *Suite) TestUpdateExpense_InvalidCategory() {
	ctx := context.Background()
	category := entity.Category("travel")

	result, err := s.uc.UpdateExpense(ctx, 7, entity.ExpensePatch{Category: &category})

	s.Equal(entity.ErrInvalidCategory, err)
	s.Zero(result)
}

func (s *Suite) TestUpdateExpense_InvalidDate() {
	ctx := context.Background()
	zero := time.Time{}

	result, err := s.uc.UpdateExpense(ctx, 7, entity.ExpensePatch{SpentAt: &zero})

	s.Equal(entity.ErrInvalidDate, err)
	s.Zero(result)
}

func (s *Suite) TestUpdateExpense_CommentTooLong() {
	ctx := context.Background()
	comment := strings.Repeat("я", 201)

	result, err := s.uc.UpdateExpense(ctx, 7, entity.ExpensePatch{Comment: &comment})

	s.Equal(entity.ErrInvalidComment, err)
	s.Zero(result)
}

func (s *Suite) TestUpdateExpense_CommentMaxLength() {
	ctx := context.Background()
	comment := strings.Repeat("я", 200)
	patch := entity.ExpensePatch{Comment: &comment}

	s.repo.EXPECT().UpdateExpense(gomock.Any(), int64(7), patch).Return(testExpense(), nil)

	_, err := s.uc.UpdateExpense(ctx, 7, patch)

	s.NoError(err)
}

func (s *Suite) TestUpdateExpense_NotFound() {
	ctx := context.Background()
	amount := int64(35000)

	s.repo.EXPECT().
		UpdateExpense(gomock.Any(), int64(999), gomock.Any()).
		Return(entity.Expense{}, entity.ErrNotFound)

	result, err := s.uc.UpdateExpense(ctx, 999, entity.ExpensePatch{Amount: &amount})

	s.ErrorIs(err, entity.ErrNotFound)
	s.Contains(err.Error(), "u.repo.UpdateExpense")
	s.Zero(result)
}

func (s *Suite) TestUpdateExpense_RepositoryError() {
	ctx := context.Background()
	repoErr := errors.New("connection refused")

	s.repo.EXPECT().
		UpdateExpense(gomock.Any(), int64(7), gomock.Any()).
		Return(entity.Expense{}, repoErr)

	result, err := s.uc.UpdateExpense(ctx, 7, entity.ExpensePatch{})

	s.Error(err)
	s.Zero(result)
	s.ErrorIs(err, repoErr)
	s.Contains(err.Error(), "u.repo.UpdateExpense")
}

func (s *Suite) TestUpdateExpense_TrimsComment() {
	ctx := context.Background()
	comment := "  " + strings.Repeat("я", 200) + "  "
	trimmed := strings.Repeat("я", 200)

	s.repo.EXPECT().
		UpdateExpense(gomock.Any(), int64(7), entity.ExpensePatch{Comment: &trimmed}).
		Return(testExpense(), nil)

	_, err := s.uc.UpdateExpense(ctx, 7, entity.ExpensePatch{Comment: &comment})

	s.NoError(err)
}
