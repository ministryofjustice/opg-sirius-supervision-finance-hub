package api

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdatePendingInvoiceAdjustment(t *testing.T) {
	logger, mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "", logger)
	r := io.NopCloser(bytes.NewReader([]byte(nil)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 204,
			Body:       r,
		}, nil
	}

	status := "approved"

	err := client.UpdatePendingInvoiceAdjustment(getContext(nil), 2, 4, status)
	assert.Equal(t, nil, err)
}

func TestUpdatePendingInvoiceAdjustmentUnauthorised(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)

	status := "approved"

	err := client.UpdatePendingInvoiceAdjustment(getContext(nil), 1, 5, status)

	assert.Equal(t, ErrUnauthorized.Error(), err.Error())
}

func TestUpdatePendingInvoiceAdjustmentReturns500Error(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)

	status := "approved"

	err := client.UpdatePendingInvoiceAdjustment(getContext(nil), 1, 2, status)
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/invoice-adjustments/2",
		Method: http.MethodPut,
	}, err)
}
