package usecase

type Usecase struct {
	repo ExpenseRepository
}

func New(repo ExpenseRepository) *Usecase {
	return &Usecase{repo: repo}
}
