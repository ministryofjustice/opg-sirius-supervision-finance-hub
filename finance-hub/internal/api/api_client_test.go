package api

import (
	"context"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/auth"
	"log/slog"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
	cache  *Caches
}

var (
	// GetDoFunc fetches the mock client's `Do` func. Implement this within a test to modify the client's behaviour.
	GetDoFunc func(req *http.Request) (*http.Response, error)
)

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

type mockJWTClient struct {
}

func (m *mockJWTClient) CreateJWT(ctx context.Context) string {
	return "jwt"
}

type mockAllPayClient struct {
	modulusCalled       bool
	modulusError        error
	createMandateCalled bool
	createMandateError  error
}

func (m *mockAllPayClient) ModulusCheck(ctx context.Context, sortCode string, accountNumber string) error {
	m.modulusCalled = true
	return m.modulusError
}

func (m *mockAllPayClient) CreateMandate(ctx context.Context, data *allpay.CreateMandateRequest) error {
	m.createMandateCalled = true
	return m.createMandateError
}

func TestClientError(t *testing.T) {
	assert.Equal(t, "message", ClientError("message").Error())
}

func TestStatusError(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/some/url", nil)

	resp := &http.Response{
		StatusCode: http.StatusTeapot,
		Request:    req,
	}

	err := newStatusError(resp)

	assert.Equal(t, "POST /some/url returned 418", err.Error())
	assert.Equal(t, err, err.Data())
}

func SetUpTest() *MockClient {
	mockClient := &MockClient{cache: newCaches()}
	return mockClient
}

func testContext() auth.Context {
	return auth.Context{
		Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-hub-test")),
	}
}

func testContextWithLogger(h slog.Handler) context.Context {
	return auth.Context{
		Context: telemetry.ContextWithLogger(context.Background(), slog.New(h)),
	}
}
