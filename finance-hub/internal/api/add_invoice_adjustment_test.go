package api

import (
	"bytes"
	"encoding/json"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddInvoiceAdjustment(t *testing.T) {
	mockClient := SetUpTest()
	mockJWT := mockJWTClient{}
	client := NewClient(mockClient, &mockJWT, Envs{"http://localhost:3000", ""}, nil)

	json := `{
	       "reference": "01234"
	   }`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 201,
			Body:       r,
		}, nil
	}

	err := client.AddInvoiceAdjustment(testContext(), 2, 41, 4, "CREDIT_MEMO", "notes here", "100", false)
	assert.Equal(t, nil, err)
}

func TestAddInvoiceAdjustmentUnauthorised(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL}, nil)

	err := client.AddInvoiceAdjustment(testContext(), 2, 41, 4, "CREDIT_MEMO", "notes here", "100", false)

	assert.Equal(t, ErrUnauthorized.Error(), err.Error())
}

func TestAddInvoiceAdjustmentReturns500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL}, nil)

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

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL}, nil)

	err := client.AddInvoiceAdjustment(testContext(), 2, 41, 4, "CREDIT_MEMO", "notes here", "100", false)
	expectedError := apierror.ValidationError{Errors: apierror.ValidationErrors{"Field": map[string]string{"Tag": "Message"}}}
	assert.Equal(t, expectedError, err.(apierror.ValidationError))
}
