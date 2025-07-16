package shared

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsSortCodeAllZeros(t *testing.T) {
	tests := []struct {
		in       string
		expected bool
	}{
		{
			in:       "22-22-22",
			expected: false,
		},
		{
			in:       "00-00-00",
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			testResult := IsSortCodeAllZeros(tt.in)
			assert.EqualValues(t, tt.expected, testResult)
		})
	}
}
