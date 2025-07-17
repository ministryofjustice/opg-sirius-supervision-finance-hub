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

func TestAddDirectDebit(t *testing.T) {
	mockClient := SetUpTest()
	mockJWT := mockJWTClient{}
	client := NewClient(mockClient, &mockJWT, Envs{"http://localhost:3000", ""})

	jsonData := `{
			"accountHolder": "CLIENT",
			"name": "Mrs Account Holder",
			"sortCode": "30-33-30",
			"account": "12345678",
        }`

	r := io.NopCloser(bytes.NewReader([]byte(jsonData)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 201,
			Body:       r,
		}, nil
	}

	err := client.AddDirectDebit(testContext(), 1, "CLIENT", "Mrs Account Holder", "30-33-30", "12345678")
	assert.Equal(t, nil, err)
}

func TestAddDirectDebitUnauthorised(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})

	err := client.AddDirectDebit(testContext(), 1, "CLIENT", "Mrs Account Holder", "30-33-30", "12345678")

	assert.Equal(t, ErrUnauthorized.Error(), err.Error())
}

func TestAddDirectDebitReturns500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})

	err := client.AddDirectDebit(testContext(), 1, "CLIENT", "Mrs Account Holder", "30-33-30", "12345678")
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/direct-debits",
		Method: http.MethodPost,
	}, err)
}

func TestAddDirectDebitReturnsValidationError(t *testing.T) {
	validationErrors := apierror.ValidationError{
		Errors: map[string]map[string]string{
			"accountHolder": {
				"required": "This field accountHolder needs to be looked at",
			},
		},
	}
	responseBody, _ := json.Marshal(validationErrors)
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write(responseBody)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})

	err := client.AddDirectDebit(testContext(), 1, "CLIENT", "Mrs Account Holder", "30-33-30", "12345678")
	expectedError := apierror.ValidationError{Errors: apierror.ValidationErrors{"accountHolder": map[string]string{"required": "This field accountHolder needs to be looked at"}}}
	assert.Equal(t, expectedError, err.(apierror.ValidationError))
}
