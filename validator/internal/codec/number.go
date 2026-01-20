// internal/codec/number.go
package codec

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func ParseInt(text string, bitSize int) (int64, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, fmt.Errorf("invalid number")
	}
	value, err := strconv.ParseInt(text, 10, bitSize)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func ParseUint(text string, bitSize int) (uint64, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, fmt.Errorf("invalid number")
	}
	value, err := strconv.ParseUint(text, 10, bitSize)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func ParseFloat(text string, bitSize int) (float64, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, fmt.Errorf("invalid number")
	}
	value, err := strconv.ParseFloat(text, bitSize)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("invalid number")
	}
	return value, nil
}
