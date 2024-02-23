package server

import (
	"github.com/opg-sirius-finance-hub/internal/model"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockInvoiceData struct {
	Invoices model.InvoiceList
	AppVars
}

func TestInvoice_error(t *testing.T) {
	client := mockApiClient{}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.SetPathValue("id", "abc")

	data := PageData{
		Data: mockInvoiceData{
			Invoices: model.InvoiceList{
				Invoices: []model.Invoice{
					{
						Id:                 3,
						Ref:                "N2000001/20",
						Status:             "Unpaid",
						Amount:             "232",
						RaisedDate:         "01/04/2222",
						Received:           "22",
						OutstandingBalance: "210",
						Ledgers: []model.Ledger{
							{
								Amount:          "123",
								ReceivedDate:    "01/05/2222",
								TransactionType: "Online card payment",
								Status:          "Applied",
							},
						},
						SupervisionLevels: []model.SupervisionLevel{
							{
								Level:  "General",
								Amount: "320",
								From:   "01/04/2019",
								To:     "31/03/2020",
							},
						},
					},
				},
			},
			AppVars: AppVars{Path: "/path/"},
		},
	}

	routeObj := route{client: client, tmpl: template, partial: "test", Data: data.Data}
	sut := InvoicesHandler{routeObj}
	err := sut.execute(w, r)

	assert.NotNil(t, err)
	assert.Equal(t, "client id in string cannot be parsed to an integer", err.Error())
}
