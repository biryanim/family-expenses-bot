package tools

type GetLastExpense struct {
	usecase Usecase
}

func NewGetLastExpense(uc Usecase) *GetLastExpense {
	return &GetLastExpense{usecase: uc}
}
