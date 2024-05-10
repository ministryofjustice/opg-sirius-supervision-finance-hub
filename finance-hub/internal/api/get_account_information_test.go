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

func TestGetAccountInformation(t *testing.T) {
	logger, mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "http://localhost:8181", logger)

	json := `{
            "outstandingBalance": 2222,
            "creditBalance": 101,
            "paymentMethod": "DEMANDED"
        }`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	expectedResponse := shared.AccountInformation{
		OutstandingBalance: 2222,
		CreditBalance:      101,
		PaymentMethod:      "DEMANDED",
	}

	headerDetails, err := client.GetAccountInformation(getContext(nil), 2)
	assert.Equal(t, expectedResponse, headerDetails)
	assert.Equal(t, nil, err)
}

func TestGetAccountInformationReturnsUnauthorisedClientError(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)
	_, err := client.GetAccountInformation(getContext(nil), 2)
	assert.Equal(t, shared.ErrUnauthorized, err)
}

func TestAccountInformationReturns500Error(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)

	_, err := client.GetAccountInformation(getContext(nil), 1)
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1",
		Method: http.MethodGet,
	}, err)
}
