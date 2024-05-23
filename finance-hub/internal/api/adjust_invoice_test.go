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

func TestAdjustInvoice(t *testing.T) {
	logger, mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "", logger)

	json := `{
            "adjustmentType": "credit write off",
            "notes": "notes here",
			"amount": "10000"
        }`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 201,
			Body:       r,
		}, nil
	}

	err := client.AdjustInvoice(getContext(nil), 2, 4, "CREDIT_MEMO", "notes here", "100")
	assert.Equal(t, nil, err)
}

func TestAdjustInvoiceUnauthorised(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)

	err := client.AdjustInvoice(getContext(nil), 2, 4, "CREDIT_MEMO", "notes here", "100")

	assert.Equal(t, ErrUnauthorized.Error(), err.Error())
}

func TestAdjustInvoiceReturns500Error(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)

	err := client.AdjustInvoice(getContext(nil), 2, 4, "CREDIT_MEMO", "notes here", "100")
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/2/invoices/4/ledger-entries",
		Method: http.MethodPost,
	}, err)
}

func TestAdjustInvoiceReturnsValidationError(t *testing.T) {
	logger, _ := SetUpTest()
	validationErrors := shared.ValidationError{
		Message: "Validation failed",
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

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)

	err := client.AdjustInvoice(getContext(nil), 2, 4, "CREDIT_MEMO", "notes here", "100")
	expectedError := shared.ValidationError{Message: "", Errors: shared.ValidationErrors{"Field": map[string]string{"Tag": "Message"}}}
	assert.Equal(t, expectedError, err.(shared.ValidationError))
}
