package api

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdatePaymentMethod(t *testing.T) {
	mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "")
	r := io.NopCloser(bytes.NewReader([]byte(nil)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 204,
			Body:       r,
		}, nil
	}

	err := client.SubmitPaymentMethod(getContext(nil), 2, "DEMANDED")
	assert.Equal(t, nil, err)
}

func TestUpdatePaymentMethodUnauthorised(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)

	err := client.SubmitPaymentMethod(getContext(nil), 1, "DEMANDED")

	assert.Equal(t, ErrUnauthorized.Error(), err.Error())
}

func TestUpdatePaymentMethodReturns500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)

	err := client.SubmitPaymentMethod(getContext(nil), 1, "DEMANDED")
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/payment-method",
		Method: http.MethodPut,
	}, err)
}
