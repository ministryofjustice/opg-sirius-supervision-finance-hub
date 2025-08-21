package api

import (
	"context"
	"log/slog"
	"net/http"
	"testing"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/auth"

	"github.com/stretchr/testify/assert"
)

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
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
	modulusCalled        bool
	modulusError         error
	createMandateCalled  bool
	createMandateError   error
	cancelMandateCalled  bool
	cancelMandateError   error
	createScheduleCalled bool
	createScheduleError  error
	data                 interface{}
}

func (m *mockAllPayClient) ModulusCheck(ctx context.Context, sortCode string, accountNumber string) error {
	m.modulusCalled = true
	return m.modulusError
}

func (m *mockAllPayClient) CreateMandate(ctx context.Context, data *allpay.CreateMandateRequest) error {
	m.createMandateCalled = true
	m.data = data
	return m.createMandateError
}

func (m *mockAllPayClient) CancelMandate(ctx context.Context, data *allpay.CancelMandateRequest) error {
	m.cancelMandateCalled = true
	m.data = data
	return m.cancelMandateError
}

func (m *mockAllPayClient) CreateSchedule(ctx context.Context, data allpay.CreateScheduleInput) error {
	m.createScheduleCalled = true
	m.data = data
	return m.createScheduleError
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
	mockClient := &MockClient{}
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
