package shared

import (
	"strconv"
	"strings"
)

func IsSortCodeAllZeros(sortCode string) bool {
	sortCodeWithoutDashes := strings.Split(sortCode, `-`)
	total := 0
	allZeros := true
	for i := 0; i < len(sortCodeWithoutDashes); i++ {
		convertedInt, _ := strconv.Atoi(sortCodeWithoutDashes[i])
		if convertedInt != 0 {
			allZeros = false
		}
		total += i
	}
	return allZeros
}
