package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/stretchr/testify/assert"
)

func TestAddInvoiceAdjustment(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/clients/2/invoices/4/invoice-adjustments":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{
	       		"reference": "01234"
	   		}`))
		case "/supervision-api/v1/tasks":
			w.WriteHeader(http.StatusCreated)
		case "/holidays.json":
			_, _ = w.Write([]byte(`{
			  "england-and-wales": {
				"division": "england-and-wales",
				"events": [
				  {
					"title": "New Year’s Day",
					"date": "2024-01-01",
					"notes": "",
					"bunting": true
				  }
				]
			  }
			}`))
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{SiriusURL: ts.URL, BackendURL: ts.URL, HolidayAPIURL: ts.URL + "/holidays.json"}, nil)

	err := client.AddInvoiceAdjustment(testContext(), 2, 41, 4, "CREDIT_MEMO", "notes here", "100", false)
	assert.Equal(t, nil, err)
}

func TestAddInvoiceAdjustmentUnauthorised(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{SiriusURL: svr.URL, BackendURL: svr.URL}, nil)

	err := client.AddInvoiceAdjustment(testContext(), 2, 41, 4, "CREDIT_MEMO", "notes here", "100", false)

	assert.Equal(t, ErrUnauthorized.Error(), err.Error())
}

func TestAddInvoiceAdjustmentReturns500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{SiriusURL: svr.URL, BackendURL: svr.URL}, nil)

	err := client.AddInvoiceAdjustment(testContext(), 2, 41, 4, "CREDIT_MEMO", "notes here", "100", false)
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/2/invoices/4/invoice-adjustments",
		Method: http.MethodPost,
	}, err)
}

func TestAddInvoiceAdjustmentReturnsValidationError(t *testing.T) {
	validationErrors := apierror.ValidationError{
		Errors: map[string]map[string]string{
			"Field": {
				"Tag": "Message",
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

	err := client.AddInvoiceAdjustment(testContext(), 2, 41, 4, "CREDIT_MEMO", "notes here", "100", false)
	expectedError := apierror.ValidationError{Errors: apierror.ValidationErrors{"Field": map[string]string{"Tag": "Message"}}}
	assert.Equal(t, expectedError, err.(apierror.ValidationError))
}
