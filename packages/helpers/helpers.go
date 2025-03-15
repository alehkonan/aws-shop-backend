package helpers

import (
	"strconv"
	"strings"
)

// Converts string price to float64
func ConvertPrice(price string) float64 {
	result, err := strconv.ParseFloat(strings.TrimSpace(price), 64)
	if err != nil {
		return 0.0
	}
	return result
}

// Converts string count to int
func ConvertCount(count string) int {
	result, err := strconv.ParseInt(count, 0, 0)
	if err != nil {
		return 0
	}
	return int(result)
}
