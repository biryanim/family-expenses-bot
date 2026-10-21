package entity

import "time"

type ExpenseFilter struct {
	UserID   *int64
	From     time.Time
	To       time.Time
	Category Category
	Limit    int
	Offset   int
}
