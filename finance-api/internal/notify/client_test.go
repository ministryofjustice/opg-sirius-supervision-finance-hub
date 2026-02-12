package notify

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

//type MockRoundTripper struct {
//	RoundTripFunc func(req *http.Request) *http.Response
//}

//func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
//	return m.RoundTripFunc(req), nil
//}

//func Test_SendEmailToNotify(t *testing.T) {
//	tests := []struct {
//		name        string
//		status      int
//		expectedErr error
//	}{
//		{
//			name:        "Status created",
//			status:      http.StatusCreated,
//			expectedErr: nil,
//		},
//		{
//			name:        "Status unauthorized",
//			status:      http.StatusUnauthorized,
//			expectedErr: apierror.Unauthorized{},
//		},
//		{
//			name:        "Status OK",
//			status:      http.StatusOK,
//			expectedErr: nil,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			mockClient := &http.Client{
//				Transport: &MockRoundTripper{
//					RoundTripFunc: func(req *http.Request) *http.Response {
//						return &http.Response{
//							StatusCode: tt.status,
//							Body:       io.NopCloser(bytes.NewReader([]byte{})),
//							Request:    req,
//						}
//					},
//				},
//			}
//			sut := Client{http: mockClient}
//			ctx := auth.Context{
//				Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test")),
//				User:    &shared.User{ID: 10},
//			}
//
//			err := sut.Send(ctx, Payload{
//				EmailAddress:    "test@email.com",
//				TemplateId:      ProcessingSuccessTemplateId,
//				Personalisation: nil,
//			})
//			assert.Equal(t, tt.expectedErr, err)
//		})
//	}
//}
