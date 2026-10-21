package config

import (
	"fmt"
	"time"
)

func CheckTimezone(tz string) error {
	if tz == "" {
		return nil
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return fmt.Errorf("time.LoadLocation: %w", err)
	}
	return nil
}
