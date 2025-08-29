package api

import (
	"bytes"
	"encoding/json"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddRefund(t *testing.T) {
	mockClient := SetUpTest()
	mockJWT := mockJWTClient{}
	client := NewClient(mockClient, &mockJWT, Envs{SiriusURL: "http://localhost:3000"}, nil)

	json := `{
			"accountName":      "Reginald Refund",
			"accountNumber":    "12345678",
			"sortCode":     	"11-22-33",
			"notes":			"this is notes",
        }`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 201,
			Body:       r,
		}, nil
	}

	err := client.AddRefund(testContext(), 1, "Reginald Refund", "12345678", "11-22-33", "this is notes")
	assert.Equal(t, nil, err)
}

func TestAddRefundUnauthorised(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{SiriusURL: svr.URL, BackendURL: svr.URL}, nil)

	err := client.AddRefund(testContext(), 1, "Reginald Refund", "12345678", "11-22-33", "")

	assert.Equal(t, ErrUnauthorized.Error(), err.Error())
}

func TestAddRefundReturns500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{SiriusURL: svr.URL, BackendURL: svr.URL}, nil)

	err := client.AddRefund(testContext(), 1, "Reginald Refund", "12345678", "11-22-33", "")
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/refunds",
		Method: http.MethodPost,
	}, err)
}

func TestAddRefundReturnsValidationError(t *testing.T) {
	validationErrors := apierror.ValidationError{
		Errors: map[string]map[string]string{
			"accountNumber": {
				"tooLong": "AccountNumber number must by 8 digits",
			},
		},
	}
	responseBody, _ := json.Marshal(validationErrors)
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write(responseBody)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{SiriusURL: svr.URL, BackendURL: svr.URL}, nil)

	err := client.AddRefund(testContext(), 1, "Reginald Refund", "12345678", "11-22-33", "")
	expectedError := apierror.ValidationError{Errors: apierror.ValidationErrors{"accountNumber": map[string]string{"tooLong": "AccountNumber number must by 8 digits"}}}
	assert.Equal(t, expectedError, err.(apierror.ValidationError))
}
