package tools

type AddExpense struct {
	usecase Usecase
}

func NewAddExpense(uc Usecase) *AddExpense {
	return &AddExpense{usecase: uc}
}
