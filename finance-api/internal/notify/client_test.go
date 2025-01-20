package notify

import (
	"bytes"
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
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

func TestServer_formatFailedLines(t *testing.T) {
	tests := []struct {
		name        string
		failedLines map[int]string
		want        []string
	}{
		{
			name:        "Empty",
			failedLines: map[int]string{},
			want:        []string(nil),
		},
		{
			name: "Unsorted lines",
			failedLines: map[int]string{
				5: "DATE_PARSE_ERROR",
				3: "CLIENT_NOT_FOUND",
				8: "DUPLICATE_PAYMENT",
				1: "DUPLICATE_PAYMENT",
			},
			want: []string{
				"Line 1: Duplicate payment line",
				"Line 3: Could not find a client with this court reference",
				"Line 5: Unable to parse date - please use the format YYYY-MM-DD HH:MM:SS",
				"Line 8: Duplicate payment line",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formattedLines := formatFailedLines(tt.failedLines)
			assert.Equal(t, tt.want, formattedLines)
		})
	}
}

type MockRoundTripper struct {
	RoundTripFunc func(req *http.Request) *http.Response
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req), nil
}

func Test_SendEmailToNotify(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		expectedErr error
	}{
		{
			name:        "Status created",
			status:      http.StatusCreated,
			expectedErr: nil,
		},
		{
			name:        "Status unauthorized",
			status:      http.StatusUnauthorized,
			expectedErr: apierror.Unauthorized{},
		},
		{
			name:        "Status OK",
			status:      http.StatusOK,
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &http.Client{
				Transport: &MockRoundTripper{
					RoundTripFunc: func(req *http.Request) *http.Response {
						return &http.Response{
							StatusCode: tt.status,
							Body:       io.NopCloser(bytes.NewReader([]byte{})),
							Request:    req,
						}
					},
				},
			}
			sut := Client{http: mockClient}
			ctx := context.Background()

			err := sut.Send(ctx, Payload{
				EmailAddress:    "test@email.com",
				TemplateId:      ProcessingSuccessTemplateId,
				Personalisation: nil,
			})
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
