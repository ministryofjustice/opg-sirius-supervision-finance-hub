package api

import (
	"bytes"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetBillingHistoryCanReturn200(t *testing.T) {
	logger, mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "", logger)

	json := `
	[
	   {
		  "clientId": 456,
		  "user": "65",
		  "date": "2099-11-04T15:04:05+00:00",
		  "event": {
			"client_id": "1",
			"type": "INVOICE_ADJUSTMENT_APPLIED",
			"adjustment_type": "Write off",
			"total": 12300,
			"payment_breakdown": {
			  "invoice_reference": {
				"id": 1,
				"reference": "S203531/19"
			  },
			  "amount": 12300
			},
			"baseBillingEvent": {
			  "type": 4
			}
		  },
		  "outstanding_balance": 0
		}
	]
	`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(rq *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	expectedResponse := []shared.BillingHistory{
		{
			User: "65",
			Date: shared.Date{Time: time.Date(2099, time.November, 4, 15, 4, 5, 0, time.UTC)},
			Event: &shared.InvoiceAdjustmentApproved{
				AdjustmentType: "Write off",
				ClientId:       "1",
				PaymentBreakdown: shared.PaymentBreakdown{
					InvoiceReference: shared.InvoiceEvent{
						ID:        1,
						Reference: "S203531/19",
					},
					Amount: 12300,
				},
				BaseBillingEvent: shared.BaseBillingEvent{Type: 4},
			},
			OutstandingBalance: 0,
		},
	}

	invoiceList, err := client.GetBillingHistory(getContext(nil), 456)

	assert.Equal(t, nil, err)
	assert.Equal(t, expectedResponse, invoiceList)
}

func TestGetBillingHistoryCanThrow500Error(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)

	_, err := client.GetBillingHistory(getContext(nil), 1)

	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/billing-history",
		Method: http.MethodGet,
	}, err)
}

func TestGetBillingHistoryUnauthorised(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)

	clientList, err := client.GetBillingHistory(getContext(nil), 3)

	var expectedResponse []shared.BillingHistory

	assert.Equal(t, expectedResponse, clientList)
	assert.Equal(t, ErrUnauthorized, err)
}
