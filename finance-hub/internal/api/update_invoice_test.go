package api

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateInvoice(t *testing.T) {
	logger, mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "", logger)

	json := `{
            "adjustmentType": "credit write off",
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

	err := client.UpdateInvoice(getContext(nil), 2, 4, "credit write off", "notes here", "100")
	assert.Equal(t, nil, err)
}

func TestUpdateInvoiceUnauthorised(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, "", logger)

	err := client.UpdateInvoice(getContext(nil), 2, 4, "credit write off", "notes here", "100")

	assert.Equal(t, ErrUnauthorized, err)
}

func TestUpdateInvoiceReturns500Error(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, "", logger)

	err := client.UpdateInvoice(getContext(nil), 2, 4, "credit write off", "notes here", "100")
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/api/v1/invoices/4/ledger-entries",
		Method: http.MethodPost,
	}, err)
}
