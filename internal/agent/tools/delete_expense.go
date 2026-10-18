package tools

type DeleteExpense struct {
	usecase Usecase
}

func NewDeleteExpense(uc Usecase) *DeleteExpense {
	return &DeleteExpense{usecase: uc}
}
