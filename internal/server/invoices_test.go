package server

import (
	"github.com/opg-sirius-finance-hub/internal/model"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInvoice_error(t *testing.T) {
	data := model.Invoices{
		model.Invoice{
			Id:                 3,
			Ref:                "N2000001/20",
			Status:             "Unpaid",
			Amount:             "232",
			RaisedDate:         model.NewDate("01/04/2222"),
			Received:           "22",
			OutstandingBalance: "210",
			Ledgers: []model.Ledger{
				{
					Amount:          "123",
					ReceivedDate:    model.NewDate("01/05/2222"),
					TransactionType: "Online card payment",
					Status:          "Applied",
				},
			},
			SupervisionLevels: []model.SupervisionLevel{
				{
					Level:  "General",
					Amount: "320",
					From:   model.NewDate("01/04/2019"),
					To:     model.NewDate("31/03/2020"),
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

	expected := InvoiceTab{
		Invoices: data,
		AppVars:  appVars,
	}

	assert.Equal(t, expected, ro.data)
}
