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

func pointer(val string) *string {
	return &val
}
func TestAddManualInvoice(t *testing.T) {
	var nilString string
	mockClient := SetUpTest()
	mockJWT := mockJWTClient{}
	client := NewClient(mockClient, &mockJWT, Envs{"http://localhost:3000", ""}, nil)

	json := `{
			"id":                "1",
			"invoiceType":       "SO",
			"amount":         	 300,
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

	err := client.AddManualInvoice(testContext(), 1, "SO", pointer("300"), pointer("05/05/2024"), &nilString, pointer("04/04/2024"), pointer("31/03/2025"), pointer("GENERAL"))
	assert.Equal(t, nil, err)
}

func TestAddManualInvoiceUnauthorised(t *testing.T) {
	var nilString string
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL}, nil)

	err := client.AddManualInvoice(testContext(), 1, "SO", pointer("2025"), pointer("05/05/2024"), &nilString, pointer("04/04/2024"), pointer("31/03/2025"), pointer("GENERAL"))

	assert.Equal(t, ErrUnauthorized.Error(), err.Error())
}

func TestAddManualInvoiceReturns500Error(t *testing.T) {
	var nilString string
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL}, nil)

	err := client.AddManualInvoice(testContext(), 1, "SO", pointer("2025"), pointer("05/05/2024"), &nilString, pointer("04/04/2024"), pointer("31/03/2025"), pointer("GENERAL"))

	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/invoices",
		Method: http.MethodPost,
	}, err)
}

func TestAddManualInvoiceReturnsBadRequestError(t *testing.T) {
	var nilString string
	mockClient := SetUpTest()
	mockJWT := mockJWTClient{}
	client := NewClient(mockClient, &mockJWT, Envs{"http://localhost:3000", ""}, nil)

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

	err := client.AddManualInvoice(testContext(), 1, "SO", pointer("2025"), pointer("05/05/2024"), &nilString, pointer("04/04/2024"), pointer("31/03/2025"), pointer("GENERAL"))

	expectedError := apierror.ValidationError{Errors: apierror.ValidationErrors{"EndDate": map[string]string{"EndDate": "EndDate"}, "StartDate": map[string]string{"StartDate": "StartDate"}}}
	assert.Equal(t, expectedError, err)
}

func TestAddManualInvoiceReturnsValidationError(t *testing.T) {
	var nilString string
	validationErrors := apierror.ValidationError{
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

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL}, nil)

	err := client.AddManualInvoice(testContext(), 1, "SO", pointer("2025"), pointer("05/05/2024"), &nilString, pointer("04/04/2024"), pointer("31/03/2025"), pointer("GENERAL"))
	expectedError := apierror.ValidationError{Errors: apierror.ValidationErrors{"DateReceived": map[string]string{"date-in-the-past": "This field DateReceived needs to be looked at date-in-the-past"}}}
	assert.Equal(t, expectedError, err.(apierror.ValidationError))
}
