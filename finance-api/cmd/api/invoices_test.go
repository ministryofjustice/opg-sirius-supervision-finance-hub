package api

import (
	"github.com/jackc/pgx/v5"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestServer_getInvoices(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1/invoices", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	dateString := "2020-03-16"
	date, _ := time.Parse("2006-01-02", dateString)

	invoicesInfo := &shared.Invoices{
		shared.Invoice{
			Id:                 1,
			Ref:                "S203531/19",
			Status:             "",
			Amount:             12,
			RaisedDate:         shared.Date{Time: date},
			Received:           123,
			OutstandingBalance: 0,
			Ledgers: []shared.Ledger{
				{
					Amount:          123,
					ReceivedDate:    shared.NewDate("11/04/2022"),
					TransactionType: "unknown",
					Status:          "Confirmed",
				},
			},
			SupervisionLevels: nil,
		},
	}

	mock := &mockService{invoices: invoicesInfo}
	server := Server{Service: mock}
	server.getInvoices(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	expected := `[{"id":1,"ref":"S203531/19","status":"","amount":12,"raisedDate":"16\/03\/2020","received":123,"outstandingBalance":0,"ledgers":[{"amount":123,"receivedDate":"11\/04\/2022","transactionType":"unknown","status":"Confirmed"}],"supervisionLevels":null}]`

	assert.Equal(t, expected, string(data))
	assert.Equal(t, 1, mock.expectedId)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getInvoices_error(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1/invoices", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	mock := &mockService{err: pgx.ErrTooManyRows}
	server := Server{Service: mock}
	server.getInvoices(w, req)

	res := w.Result()

	assert.Equal(t, 500, res.StatusCode)
}
