package api

import (
	"bytes"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdatePaymentMethod(t *testing.T) {
	mockClient := SetUpTest()
	mockJWT := mockJWTClient{}
	client := NewClient(mockClient, &mockJWT, Envs{"http://localhost:3000", ""}, nil)
	r := io.NopCloser(bytes.NewReader([]byte(nil)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 204,
			Body:       r,
		}, nil
	}

	err := client.UpdatePaymentMethod(testContext(), 2, shared.PaymentMethodDemanded.Key())
	assert.Equal(t, nil, err)
}

func TestUpdatePaymentMethodUnauthorised(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL}, nil)

	err := client.UpdatePaymentMethod(testContext(), 1, shared.PaymentMethodDemanded.Key())

	assert.Equal(t, ErrUnauthorized.Error(), err.Error())
}

func TestUpdatePaymentMethodReturns500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL}, nil)

	err := client.UpdatePaymentMethod(testContext(), 1, shared.PaymentMethodDemanded.Key())
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/payment-method",
		Method: http.MethodPut,
	}, err)
}
