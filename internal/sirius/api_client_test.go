package sirius

import (
	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/opg-sirius-supervision-finance-hub/internal/mocks"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func SetUpTest() (*logging.Logger, *mocks.MockClient) {
	logger := logging.New(os.Stdout, "opg-sirius-supervision-finance-hub")
	mockClient := &mocks.MockClient{}
	return logger, mockClient
}
