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

func TestServer_getInvoices(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1/invoices", nil)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()
	dateString := "2020-03-16"
	date, _ := time.Parse("2006-01-02", dateString)

	invoicesInfo := shared.Invoices{
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
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	_ = server.getInvoices(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	expected := `[{"id":1,"ref":"S203531/19","status":"","amount":12,"raisedDate":"16\/03\/2020","received":123,"outstandingBalance":0,"ledgers":[{"amount":123,"receivedDate":"11\/04\/2022","transactionType":"unknown","status":"Confirmed"}],"supervisionLevels":null}]`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(w.Body.String()))
	assert.Equal(t, 1, mock.expectedIds[0])
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getInvoices_returns_an_empty_array(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/2/invoices", nil)
	req.SetPathValue("clientId", "2")
	w := httptest.NewRecorder()

	invoicesInfo := shared.Invoices{}

	mock := &mockService{invoices: invoicesInfo}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	_ = server.getInvoices(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	expected := `[]`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(w.Body.String()))
	assert.Equal(t, 2, mock.expectedIds[0])
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getInvoices_error(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1/invoices", nil)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	mock := &mockService{errs: map[string]error{"GetInvoices": pgx.ErrTooManyRows}}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	err := server.getInvoices(w, req)

	assert.Error(t, err)
}
