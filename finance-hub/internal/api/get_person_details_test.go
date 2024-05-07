package api

import (
	"bytes"
	"github.com/opg-sirius-finance-hub/finance-hub/internal/config"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPersonDetails(t *testing.T) {
	logger, mockClient, envVars := SetUpTest()
	client, _ := NewApiClient(mockClient, logger, envVars)

	json := `{
            "id": 2,
            "firstname": "Finance",
            "surname": "Person",
			"caseRecNumber": "12345678"
        }`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	expectedResponse := shared.Person{
		ID:        2,
		FirstName: "Finance",
		Surname:   "Person",
		CourtRef:  "12345678",
	}

	person, err := client.GetPersonDetails(getContext(nil), 2)
	assert.Equal(t, expectedResponse, person)
	assert.Equal(t, nil, err)
}

func TestGetPersonDetailsReturnsUnauthorisedClientError(t *testing.T) {
	logger, _, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	envVars := config.EnvironmentVars{SiriusURL: svr.URL, BackendURL: svr.URL}
	client, _ := NewApiClient(http.DefaultClient, logger, envVars)

	_, err := client.GetPersonDetails(getContext(nil), 2)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestPersonDetailsReturns500Error(t *testing.T) {
	logger, _, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	envVars := config.EnvironmentVars{SiriusURL: svr.URL, BackendURL: svr.URL}
	client, _ := NewApiClient(http.DefaultClient, logger, envVars)

	_, err := client.GetPersonDetails(getContext(nil), 1)
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/api/v1/clients/1",
		Method: http.MethodGet,
	}, err)
}
