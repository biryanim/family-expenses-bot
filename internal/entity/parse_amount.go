package entity

import (
	"math"
	"strconv"
	"strings"
)

func ParseAmount(s string) (int64, error) {
	rublesPart, kopecksPart, hasFraction := strings.Cut(s, ".")
	if rublesPart == "" || (hasFraction && (kopecksPart == "" || len(kopecksPart) > 2)) {
		return 0, ErrInvalidAmount
	}
	for _, r := range rublesPart + kopecksPart {
		if r < '0' || r > '9' {
			return 0, ErrInvalidAmount
		}
	}

	rubles, err := strconv.ParseInt(rublesPart, 10, 64)
	if err != nil {
		return 0, ErrInvalidAmount
	}
	var kopecks int64
	if hasFraction {
		kopecks, err = strconv.ParseInt(kopecksPart, 10, 64)
		if err != nil {
			return 0, ErrInvalidAmount
		}
		if len(kopecksPart) == 1 {
			kopecks *= 10
		}
	}

	if rubles > (math.MaxInt64-kopecks)/100 {
		return 0, ErrInvalidAmount
	}
	amount := rubles*100 + kopecks
	if amount <= 0 {
		return 0, ErrInvalidAmount
	}
	return amount, nil
}
