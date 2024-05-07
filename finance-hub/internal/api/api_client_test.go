package api

import (
	"context"
	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/opg-sirius-finance-hub/finance-hub/internal/config"
	"net/http"
	"os"
	"testing"

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

func getContext(cookies []*http.Cookie) Context {
	return Context{
		Context:   context.Background(),
		Cookies:   cookies,
		XSRFToken: "abcde",
	}
}

func TestClientError(t *testing.T) {
	assert.Equal(t, "message", ClientError("message").Error())
}

func TestValidationError(t *testing.T) {
	assert.Equal(t, "message", ValidationError{Message: "message"}.Error())
}

func TestStatusError(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/some/url", nil)

	resp := &http.Response{
		StatusCode: http.StatusTeapot,
		Request:    req,
	}

	err := newStatusError(resp)

	assert.Equal(t, "POST /some/url returned 418", err.Error())
	assert.Equal(t, "unexpected response from Sirius", err.Title())
	assert.Equal(t, err, err.Data())
}

func SetUpTest() (*logging.Logger, *MockClient, config.EnvironmentVars) {
	logger := logging.New(os.Stdout, "opg-sirius-finance-hub")
	mockClient := &MockClient{}
	envVars := config.EnvironmentVars{
		SiriusURL:  "http://localhost:3000",
		BackendURL: "http://localhost:8181",
	}
	return logger, mockClient, envVars
}
