package shared

import (
	"fmt"
	"strconv"
	"strings"
)

func IntToDecimalString(i int) string {
	whole := i / 100
	fraction := i % 100
	if fraction == 0 {
		return fmt.Sprintf("%d", whole)
	}
	return fmt.Sprintf("%d.%02d", whole, fraction)
}

func DecimalStringToInt(s string) int32 {
	parts := strings.SplitN(s, ".", 2)

	// Ensure two decimal places
	if len(parts) == 1 {
		parts = append(parts, "00")
	} else if len(parts[1]) == 1 {
		parts[1] += "0"
	} else if len(parts[1]) > 2 {
		parts[1] = parts[1][:2] // truncate extra decimals
	}

	combined := parts[0] + parts[1]
	val, _ := strconv.ParseInt(combined, 10, 32)
	return int32(val)
}
