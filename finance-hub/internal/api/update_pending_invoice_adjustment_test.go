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
	mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "")
	r := io.NopCloser(bytes.NewReader([]byte(nil)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 204,
			Body:       r,
		}, nil
	}

	err := client.UpdatePendingInvoiceAdjustment(getContext(nil), 2, 4, "APPROVED")
	assert.Equal(t, nil, err)
}

func TestUpdatePendingInvoiceAdjustmentUnauthorised(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)

	err := client.UpdatePendingInvoiceAdjustment(getContext(nil), 1, 5, "APPROVED")

	assert.Equal(t, ErrUnauthorized.Error(), err.Error())
}

func TestUpdatePendingInvoiceAdjustmentReturns500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)

	err := client.UpdatePendingInvoiceAdjustment(getContext(nil), 1, 2, "APPROVED")
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/invoice-adjustments/2",
		Method: http.MethodPut,
	}, err)
}
