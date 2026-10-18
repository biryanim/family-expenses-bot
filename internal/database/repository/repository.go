package repository

const expenseColumns = "id, user_id, amount, category, spent_at, comment, created_at, updated_at"

const (
	constraintExpenseUserFK         = "expenses_user_id_fkey"
	constraintExpenseAmountPositive = "expenses_amount_positive"
	constraintExpenseCategoryValid  = "expenses_category_valid"
)

type Repository struct {
	pool Pool
}

func New(pool Pool) *Repository {
	return &Repository{pool: pool}
}
