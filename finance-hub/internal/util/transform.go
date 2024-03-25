package util

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
