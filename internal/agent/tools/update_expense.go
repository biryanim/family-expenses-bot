package tools

type UpdateExpense struct {
	usecase Usecase
}

func NewUpdateExpense(uc Usecase) *UpdateExpense {
	return &UpdateExpense{usecase: uc}
}
