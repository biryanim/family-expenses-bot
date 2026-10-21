package entity

import "time"

type ExpensePatch struct {
	Amount   *int64
	Category *Category
	SpentAt  *time.Time
	Comment  *string
}
