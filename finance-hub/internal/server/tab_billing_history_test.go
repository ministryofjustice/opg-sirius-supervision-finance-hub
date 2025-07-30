package server

import (
	"context"
	"errors"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
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
				InvoiceType:      shared.InvoiceTypeAD,
				Amount:           65498,
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceGenerated},
			},
			OutstandingBalance: 25124,
		},
	}

	client := mockApiClient{
		BillingHistory: data,
		User:           shared.User{DisplayName: "Mr Testman"},
	}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	ctx := telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("opg-sirius-supervision-finance-hub"))
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "test-url/1", nil)
	r.SetPathValue("clientId", "456")

	appVars := AppVars{Path: "/path/"}

	sut := BillingHistoryHandler{ro}
	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.True(t, ro.executed)

	expected := &BillingHistoryTab{
		BillingHistory: []BillingHistory{
			{
				User:               "Mr Testman",
				Date:               data[0].Date,
				Event:              data[0].Event,
				OutstandingBalance: "251.24",
				CreditBalance:      "0",
			},
		},
		AppVars: appVars,
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
