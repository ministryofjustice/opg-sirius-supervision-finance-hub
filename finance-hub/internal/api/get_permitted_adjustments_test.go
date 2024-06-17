package api

import (
	"bytes"
	"github.com/opg-sirius-finance-hub/shared"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPermittedAdjustments(t *testing.T) {
	logger, mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "http://localhost:8181", logger)

	json := `{["CREDIT MEMO","DEBIT MEMO"]}`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	expectedResponse := []shared.AdjustmentType{shared.AdjustmentTypeAddCredit, shared.AdjustmentTypeAddDebit}

	types, err := client.GetPermittedAdjustments(getContext(nil), 1, 2)
	assert.Equal(t, expectedResponse, types)
	assert.Equal(t, nil, err)
}

func TestGetPermittedAdjustmentsReturnsUnauthorisedClientError(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)
	_, err := client.GetPermittedAdjustments(getContext(nil), 1, 2)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestGetPermittedAdjustmentsReturns500Error(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)

	_, err := client.GetPermittedAdjustments(getContext(nil), 1, 2)
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/invoices/2/permitted-adjustments",
		Method: http.MethodGet,
	}, err)
}
