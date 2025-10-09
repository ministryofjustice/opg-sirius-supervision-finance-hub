package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/auth"

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
