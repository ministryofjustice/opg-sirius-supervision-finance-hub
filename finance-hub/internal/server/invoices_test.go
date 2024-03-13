package server

import (
	"errors"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInvoice(t *testing.T) {
	data := shared.Invoices{
		shared.Invoice{
			Id:                 3,
			Ref:                "N2000001/20",
			Status:             "Unpaid",
			Amount:             "232",
			RaisedDate:         shared.NewDate("01/04/2222"),
			Received:           "22",
			OutstandingBalance: "210",
			Ledgers: []shared.Ledger{
				{
					Amount:          "123",
					ReceivedDate:    shared.NewDate("01/05/2222"),
					TransactionType: "Online card payment",
					Status:          "Applied",
				},
			},
			SupervisionLevels: []shared.SupervisionLevel{
				{
					Level:  "General",
					Amount: "320",
					From:   shared.NewDate("01/04/2019"),
					To:     shared.NewDate("31/03/2020"),
				},
			},
		},
	}

	client := mockApiClient{Invoices: data}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.SetPathValue("id", "1")

	appVars := AppVars{Path: "/path/"}

	sut := InvoicesHandler{ro}
	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.True(t, ro.executed)

	expected := &InvoiceTab{
		Invoices: data,
		AppVars:  appVars,
	}

	assert.Equal(t, expected, ro.data)
}

func TestInvoiceErrors(t *testing.T) {
	client := mockApiClient{}
	client.error = errors.New("this has failed")
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.SetPathValue("id", "1")

	appVars := AppVars{Path: "/path/"}

	sut := InvoicesHandler{ro}
	err := sut.render(appVars, w, r)

	assert.Equal(t, "this has failed", err.Error())
	assert.False(t, ro.executed)
}
