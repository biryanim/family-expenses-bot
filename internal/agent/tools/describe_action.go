package tools

func (e Expense) DescribeAction() string {
	return "Трата на " + e.Amount + " ₽ за " + e.Date
}
