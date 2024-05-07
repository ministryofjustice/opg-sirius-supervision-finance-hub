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

func TestServer_getInvoiceAdjustments(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1/invoice-adjustments", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	dateString := "2020-03-16"
	date, _ := time.Parse("2006-01-02", dateString)

	ia := &shared.InvoiceAdjustments{
		shared.InvoiceAdjustment{
			Id:             1,
			InvoiceRef:     "abc123/24",
			RaisedDate:     shared.Date{Time: date},
			AdjustmentType: shared.AdjustmentTypeAddCredit,
			Amount:         123400,
			Status:         "PENDING",
			Notes:          "Credit memo for invoice",
		},
	}

	mock := &mockService{invoiceAdjustments: ia}
	server := Server{Service: mock}
	server.getInvoiceAdjustments(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	expected := `[{"id":1,"invoiceRef":"abc123/24","raisedDate":"16\/03\/2020","adjustmentType":"CREDIT_MEMO","amount":123400,"status":"PENDING","notes":"Credit memo for invoice"}]`

	assert.Equal(t, expected, string(data))
	assert.Equal(t, 1, mock.expectedIds[0])
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getInvoiceAdjustments_returns_an_empty_array(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/2/invoice-adjustments", nil)
	req.SetPathValue("id", "2")
	w := httptest.NewRecorder()

	ia := &shared.InvoiceAdjustments{}

	mock := &mockService{invoiceAdjustments: ia}
	server := Server{Service: mock}
	server.getInvoiceAdjustments(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	expected := `[]`

	assert.Equal(t, expected, string(data))
	assert.Equal(t, 2, mock.expectedIds[0])
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getInvoiceAdjustments_error(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1/invoice-adjustments", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	mock := &mockService{err: pgx.ErrTooManyRows}
	server := Server{Service: mock}
	server.getInvoiceAdjustments(w, req)

	res := w.Result()

	assert.Equal(t, 500, res.StatusCode)
}
