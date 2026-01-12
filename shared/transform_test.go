package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IntToCurrency(t *testing.T) {
	tests := []struct {
		name string
		arg  int
		want string
	}{
		{
			"converts int to two decimal places",
			12345,
			"£123.45",
		},
		{
			"displays two decimal places when the last digit is 0",
			12340,
			"£123.40",
		},
		{
			"displays no decimal places when the last two digits are 0",
			12300,
			"£123",
		},
		{
			"displays a leading zero and two decimal places when there are two or fewer digits",
			12,
			"£0.12",
		},
		{
			"negative value",
			-1200,
			"-£12",
		},
		{
			"negative value with decimals",
			-1234,
			"-£12.34",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IntToCurrency(tt.arg), "IntToCurrency(%v)", tt.arg)
		})
	}
}

func TestDecimalStringToInt(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want int32
	}{
		{
			"converts two decimal places to string",
			"123.45",
			12345,
		},
		{
			"converts trailing zero",
			"123.40",
			12340,
		},
		{
			"converts no decimals",
			"123",
			12300,
		},
		{
			"converts leading zero",
			"0.12",
			12,
		},
		{
			"floating point conversion",
			"134.14",
			13414,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, DecimalStringToInt(tt.arg), "DecimalStringToInt(%v)", tt.arg)
		})
	}
}
