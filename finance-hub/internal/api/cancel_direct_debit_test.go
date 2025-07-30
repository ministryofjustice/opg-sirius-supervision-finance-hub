package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCancelDirectDebit(t *testing.T) {
	mockClient := SetUpTest()
	mockJWT := mockJWTClient{}
	client := NewClient(mockClient, &mockJWT, Envs{"http://localhost:3000", ""})

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 202,
			Body:       nil,
		}, nil
	}

	err := client.CancelDirectDebit(testContext(), 1)
	assert.Equal(t, nil, err)
}

func TestCancelDirectDebitUnauthorised(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})

	err := client.CancelDirectDebit(testContext(), 1)

	assert.Equal(t, ErrUnauthorized.Error(), err.Error())
}

func TestCancelDirectDebitReturns500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})

	err := client.CancelDirectDebit(testContext(), 1)
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/direct-debits",
		Method: http.MethodDelete,
	}, err)
}
