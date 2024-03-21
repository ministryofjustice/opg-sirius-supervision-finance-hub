package api

import (
	"bytes"
	"github.com/opg-sirius-finance-hub/finance-hub/internal/mocks"
	"github.com/opg-sirius-finance-hub/shared"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHeaderAccountDetails(t *testing.T) {
	logger, mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "http://localhost:8181", logger)

	json := `{
            "clientId": 2,
            "cachedOutstandingBalance": "22.22",
            "cachedCreditBalance": "1.01",
            "paymentMethod": "Demanded"
        }`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	expectedResponse := shared.AccountInformation{
		OutstandingBalance: 2222,
		CreditBalance:      101,
		PaymentMethod:      "Demanded",
	}

	headerDetails, err := client.GetAccountInformation(getContext(nil), 2)
	assert.Equal(t, expectedResponse, headerDetails)
	assert.Equal(t, nil, err)
}

func TestGetHeaderAccountDetailsReturnsUnauthorisedClientError(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)
	_, err := client.GetAccountInformation(getContext(nil), 2)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestHeaderAccountDetailsReturns500Error(t *testing.T) {
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
