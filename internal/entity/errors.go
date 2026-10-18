package entity

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrUnknownUser     = errors.New("unknown family member")
	ErrInvalidAmount   = errors.New("invalid amount")
	ErrInvalidCategory = errors.New("invalid category")
	ErrInvalidDate     = errors.New("invalid date")
)
