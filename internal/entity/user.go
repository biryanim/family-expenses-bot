package entity

import "time"

type User struct {
	TelegramID int64
	Username   string
	FirstName  string
	LastName   string
	CreatedAt  time.Time
	UpdatedAt  *time.Time
}
