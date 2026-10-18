package entity

import "fmt"

func FormatAmount(kopecks int64) string {
	if kopecks%100 == 0 {
		return fmt.Sprintf("%d", kopecks/100)
	}
	return fmt.Sprintf("%d.%02d", kopecks/100, kopecks%100)
}
