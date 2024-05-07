package api

import (
	"bytes"
	"github.com/opg-sirius-finance-hub/finance-hub/internal/config"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateInvoice(t *testing.T) {
	logger, mockClient, envVars := SetUpTest()
	client, _ := NewApiClient(mockClient, logger, envVars)

	json := `{
            "invoiceType": "writeOff",
            "notes": "notes here",
			"amount": "100"
        }`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 201,
			Body:       r,
		}, nil
	}

	err := client.UpdateInvoice(getContext(nil), 2, 4, "writeOff", "notes here", "100")
	assert.Equal(t, nil, err)
}

func TestUpdateInvoiceUnauthorised(t *testing.T) {
	logger, _, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	envVars := config.EnvironmentVars{SiriusURL: svr.URL, BackendURL: svr.URL}
	client, _ := NewApiClient(http.DefaultClient, logger, envVars)

	err := client.UpdateInvoice(getContext(nil), 2, 4, "writeOff", "notes here", "100")

	assert.Equal(t, ErrUnauthorized, err)
}

func TestUpdateInvoiceReturns500Error(t *testing.T) {
	logger, _, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	envVars := config.EnvironmentVars{SiriusURL: svr.URL, BackendURL: svr.URL}
	client, _ := NewApiClient(http.DefaultClient, logger, envVars)

	err := client.UpdateInvoice(getContext(nil), 2, 4, "writeOff", "notes here", "100")
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/api/v1/invoices/4/ledger-entries",
		Method: http.MethodPost,
	}, err)
}
