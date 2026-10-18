package entity

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrUnknownUser     = errors.New("unknown family member")
	ErrInvalidAmount   = errors.New("invalid amount")
	ErrInvalidCategory = errors.New("invalid category")
	ErrInvalidDate     = errors.New("invalid date")
	ErrEmptyCompletion = errors.New("llm response has no choices")
	ErrAgentStepLimit  = errors.New("agent step limit exceeded")
	ErrEmptyAgentReply = errors.New("agent reply is empty")
)
