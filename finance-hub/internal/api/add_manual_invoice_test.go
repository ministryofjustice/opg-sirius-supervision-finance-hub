package api

import (
	"bytes"
	"encoding/json"
	"github.com/opg-sirius-finance-hub/shared"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddManualInvoice(t *testing.T) {
	mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "")

	json := `{
			"id":                "1",
			"invoiceTpe":        "SO",
			"amount":         	 2025,
			"raisedDate":    	 "05/05/2024",
			"startDate":         "01/04/2024",
			"endDate":           "31/03/2025",
			"supervisionLevel":  "GENERAL",
        }`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 201,
			Body:       r,
		}, nil
	}

	err := client.AddManualInvoice(getContext(nil), 1, "SO", "2025", "05/05/2024", "04/04/2024", "31/03/2025", "GENERAL")
	assert.Equal(t, nil, err)
}

func TestAddManualInvoiceUnauthorised(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)

	err := client.AddManualInvoice(getContext(nil), 1, "SO", "2025", "05/05/2024", "04/04/2024", "31/03/2025", "GENERAL")

	assert.Equal(t, ErrUnauthorized.Error(), err.Error())
}

func TestAddManualInvoiceReturns500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)

	err := client.AddManualInvoice(getContext(nil), 1, "SO", "2025", "05/05/2024", "04/04/2024", "31/03/2025", "GENERAL")

	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/invoices",
		Method: http.MethodPost,
	}, err)
}

func TestAddManualInvoiceReturnsBadRequestError(t *testing.T) {
	mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "")

	json := `
		{"reasons":["StartDate","EndDate"]}
	`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 400,
			Body:       r,
		}, nil
	}

	err := client.AddManualInvoice(getContext(nil), 1, "SO", "2025", "05/05/2024", "04/04/2024", "31/03/2025", "GENERAL")

	expectedError := shared.ValidationError{Message: "", Errors: shared.ValidationErrors{"EndDate": map[string]string{"EndDate": "EndDate"}, "StartDate": map[string]string{"StartDate": "StartDate"}}}
	assert.Equal(t, expectedError, err)
}

func TestAddManualInvoiceReturnsValidationError(t *testing.T) {
	validationErrors := shared.ValidationError{
		Message: "Validation failed",
		Errors: map[string]map[string]string{
			"DateReceived": {
				"date-in-the-past": "This field DateReceived needs to be looked at date-in-the-past",
			},
		},
	}
	responseBody, _ := json.Marshal(validationErrors)
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write(responseBody)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)

	err := client.AddManualInvoice(getContext(nil), 1, "SO", "2025", "05/05/9999", "04/04/2024", "31/03/2025", "GENERAL")
	expectedError := shared.ValidationError{Message: "", Errors: shared.ValidationErrors{"DateReceived": map[string]string{"date-in-the-past": "This field DateReceived needs to be looked at date-in-the-past"}}}
	assert.Equal(t, expectedError, err.(shared.ValidationError))
}
