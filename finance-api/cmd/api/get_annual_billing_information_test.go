package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func TestServer_getAnnualBillingInformation(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/annual-billing-letters", nil)
	w := httptest.NewRecorder()

	billingInfo := shared.AnnualBillingInformation{
		AnnualBillingYear:        "2023",
		DemandedExpectedCount:    6,
		DemandedIssuedCount:      12,
		DemandedSkippedCount:     1,
		DirectDebitExpectedCount: 4,
		DirectDebitIssuedCount:   2,
	}

	mock := &mockService{annualBillingInformation: billingInfo}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	_ = server.getAnnualBillingInformation(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	expected := `{"AnnualBillingYear":"2023","DemandedExpectedCount":6,"DemandedIssuedCount":12,"DemandedSkippedCount":1,"DirectDebitExpectedCount":4,"DirectDebitIssuedCount":2}`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(w.Body.String()))
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getAnnualBillingInformation_returnsEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/annual-billing-letters", nil)
	w := httptest.NewRecorder()

	billingInfo := shared.AnnualBillingInformation{}

	mock := &mockService{annualBillingInformation: billingInfo}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	_ = server.getAnnualBillingInformation(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	expected := `{"AnnualBillingYear":"","DemandedExpectedCount":0,"DemandedIssuedCount":0,"DemandedSkippedCount":0,"DirectDebitExpectedCount":0,"DirectDebitIssuedCount":0}`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(w.Body.String()))
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getAnnualBillingInformation_error(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/annual-billing-letters", nil)
	w := httptest.NewRecorder()

	mock := &mockService{errs: map[string]error{"GetAnnualBillingInformation": pgx.ErrTooManyRows}}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	err := server.getAnnualBillingInformation(w, req)

	assert.Error(t, err)
}
