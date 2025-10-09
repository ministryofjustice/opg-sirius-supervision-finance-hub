package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func TestServer_getBillingHistory(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1/billing-history", nil)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	billingHistoryInfo := []shared.BillingHistory{
		{
			User: 2,
			Date: shared.Date{Time: time.Date(2099, time.November, 4, 15, 4, 5, 0, time.UTC)},
			Event: &shared.InvoiceAdjustmentPending{
				AdjustmentType: shared.AdjustmentTypeWriteOff,
				ClientId:       1,
				PaymentBreakdown: shared.PaymentBreakdown{
					InvoiceReference: shared.InvoiceEvent{
						ID:        1,
						Reference: "S203531/19",
					},
					Amount: 12300,
					Status: "",
				},
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceAdjustmentPending},
			},
			OutstandingBalance: 0,
		},
	}

	mock := &mockService{billingHistory: billingHistoryInfo}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	_ = server.getBillingHistory(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	expected := `[{"user":2,"date":"04\/11\/2099","event":{"adjustment_type":"CREDIT WRITE OFF","client_id":1,"notes":"","payment_breakdown":{"invoice_reference":{"id":1,"reference":"S203531/19"},"amount":12300,"status":""},"type":"INVOICE_ADJUSTMENT_PENDING"},"outstanding_balance":0,"credit_balance":0}]`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(w.Body.String()))
	assert.Equal(t, 1, mock.expectedIds[0])
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getBillingHistory_returns_an_empty_array(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/2/billing-history", nil)
	req.SetPathValue("clientId", "2")
	w := httptest.NewRecorder()

	mock := &mockService{billingHistory: []shared.BillingHistory{}}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	_ = server.getBillingHistory(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	expected := `[]`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(w.Body.String()))
	assert.Equal(t, 2, mock.expectedIds[0])
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getBillingHistory_error(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1/billing-history", nil)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	mock := &mockService{errs: map[string]error{"GetBillingHistory": pgx.ErrTooManyRows}}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	err := server.getBillingHistory(w, req)

	assert.Error(t, err)
}
