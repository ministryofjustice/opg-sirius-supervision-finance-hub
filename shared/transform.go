package shared

import (
	"regexp"
	"strconv"
)

func IntToDecimalString(i int) string {
	s := strconv.FormatFloat(float64(i)/100, 'f', -1, 32)
	const singleDecimal = "\\.\\d$"
	m, _ := regexp.Match(singleDecimal, []byte(s))
	if m {
		s = s + "0"
	}
	return s
}

func DecimalStringToInt(s string) int32 {
	i, _ := strconv.ParseFloat(s, 32)
	return int32(i * 100)
}
