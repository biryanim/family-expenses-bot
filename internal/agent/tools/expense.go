package tools

import "github.com/biryanim/family-expenses-bot/internal/entity"

type Expense struct {
	ID       int64           `json:"id"`
	Amount   string          `json:"amount"`
	Category entity.Category `json:"category"`
	Date     string          `json:"date"`
	Comment  string          `json:"comment"`
}

type ExpenseList struct {
	Count     int       `json:"count"`
	Total     string    `json:"total"`
	Expenses  []Expense `json:"expenses"`
	Truncated bool      `json:"truncated"`
}
