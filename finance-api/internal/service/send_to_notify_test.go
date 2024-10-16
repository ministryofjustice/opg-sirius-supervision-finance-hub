package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_parseNotifyApiKey(t *testing.T) {
	tests := []struct {
		name             string
		key              string
		expectedIss      string
		expectedJwtToken string
	}{
		{
			name:             "Empty API key",
			key:              "",
			expectedIss:      "",
			expectedJwtToken: "",
		},
		{
			name:             "API key with too many dashes",
			key:              "oh-no-1234abcd-1234-abcd-5678-123456abcdef-hehe0101-asdf-1234-hehe-12345678abcd",
			expectedIss:      "",
			expectedJwtToken: "",
		},
		{
			name:             "Normal shaped API key",
			key:              "hehe-1234abcd-1234-abcd-5678-123456abcdef-hehe0101-asdf-1234-hehe-12345678abcd",
			expectedIss:      "1234abcd-1234-abcd-5678-123456abcdef",
			expectedJwtToken: "hehe0101-asdf-1234-hehe-12345678abcd",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iss, jwtToken := parseNotifyApiKey(tt.key)
			assert.Equal(t, tt.expectedIss, iss)
			assert.Equal(t, tt.expectedJwtToken, jwtToken)
		})
	}
}
