package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_intToDecimalString(t *testing.T) {
	type args struct {
		i int
	}
	tests := []struct {
		name string
		arg  int
		want string
	}{
		{
			"converts int to two decimal places",
			12345,
			"123.45",
		},
		{
			"displays two decimal places when the last digit is 0",
			12340,
			"123.40",
		},
		{
			"displays no decimal places when the last two digits are 0",
			12300,
			"123",
		},
		{
			"displays a leading zero and two decimal places when there are two or fewer digits",
			12,
			"0.12",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, intToDecimalString(tt.arg), "intToDecimalString(%v)", tt.arg)
		})
	}
}
