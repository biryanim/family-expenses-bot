package tools

type ListExpenses struct {
	usecase Usecase
}

func NewListExpenses(uc Usecase) *ListExpenses {
	return &ListExpenses{usecase: uc}
}
