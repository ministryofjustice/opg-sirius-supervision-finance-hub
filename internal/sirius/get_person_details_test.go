package sirius

import (
	"bytes"
	"github.com/opg-sirius-finance-hub/internal/mocks"
	"github.com/opg-sirius-finance-hub/internal/model"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPersonDetails(t *testing.T) {
	logger, mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", logger)

	json := `{
            "id": 2,
            "firstname": "Finance",
            "surname": "Person",
            "caseRecNumber": "12345678",
            "outstandingBalance": "£22.22",
            "creditBalance": "£1.01",
            "paymentMethod": "Demanded"
        }`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	expectedResponse := model.Person{
		ID:                 2,
		Firstname:          "Finance",
		Surname:            "Person",
		CourtRef:           "12345678",
		OutstandingBalance: "£22.22",
		CreditBalance:      "£1.01",
		PaymentMethod:      "Demanded",
	}

	person, err := client.GetPersonDetails(getContext(nil), 2)
	assert.Equal(t, expectedResponse, person)
	assert.Equal(t, nil, err)
}

func TestGetPersonDetailsReturnsUnauthorisedClientError(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, logger)
	_, err := client.GetPersonDetails(getContext(nil), 2)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestPersonDetailsReturns500Error(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, logger)

	_, err := client.GetPersonDetails(getContext(nil), 1)
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/api/v1/clients/1",
		Method: http.MethodGet,
	}, err)
}
