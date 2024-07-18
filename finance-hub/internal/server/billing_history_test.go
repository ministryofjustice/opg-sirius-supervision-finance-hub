package server

import (
	"errors"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBillingHistory(t *testing.T) {
	data := []shared.BillingHistory{
		{
			User: 1,
			Date: shared.NewDate("01/04/2222"),
			Event: shared.InvoiceGenerated{
				ClientId: 456,
				InvoiceReference: shared.InvoiceEvent{
					ID:        1,
					Reference: "A12345678/02",
				},
				InvoiceType:      "AD",
				InvoiceName:      "The name of the invoice",
				Amount:           65498,
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceGenerated},
			},
			OutstandingBalance: 25124,
		},
	}

	client := mockApiClient{BillingHistory: data}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.SetPathValue("clientId", "456")

	appVars := AppVars{Path: "/path/"}

	sut := BillingHistoryHandler{ro}
	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.True(t, ro.executed)

	expected := &BillingHistoryVars{
		BillingHistory: data,
		AppVars:        appVars,
	}

	assert.Equal(t, expected, ro.data)
}

func TestBillingHistoryErrors(t *testing.T) {
	client := mockApiClient{}
	client.error = errors.New("this has failed")
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.SetPathValue("clientId", "1")

	appVars := AppVars{Path: "/path/"}

	sut := BillingHistoryHandler{ro}
	err := sut.render(appVars, w, r)

	assert.Equal(t, "this has failed", err.Error())
	assert.False(t, ro.executed)
}
