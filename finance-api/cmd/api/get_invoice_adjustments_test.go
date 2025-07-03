package api

import (
	"github.com/jackc/pgx/v5"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestServer_getInvoiceAdjustments(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1/invoice-adjustments", nil)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()
	dateString := "2020-03-16"
	date, _ := time.Parse("2006-01-02", dateString)

	ia := shared.InvoiceAdjustments{
		shared.InvoiceAdjustment{
			Id:             1,
			InvoiceRef:     "abc123/24",
			RaisedDate:     shared.Date{Time: date},
			AdjustmentType: shared.AdjustmentTypeCreditMemo,
			Amount:         123400,
			Status:         "PENDING",
			Notes:          "Credit memo for invoice",
			CreatedBy:      2,
		},
	}

	mock := &mockService{invoiceAdjustments: ia}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	_ = server.getInvoiceAdjustments(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	expected := `[{"id":1,"invoiceRef":"abc123/24","raisedDate":"16\/03\/2020","adjustmentType":"CREDIT MEMO","amount":123400,"status":"PENDING","notes":"Credit memo for invoice","createdBy":2}]`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(w.Body.String()))
	assert.Equal(t, 1, mock.expectedIds[0])
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getInvoiceAdjustments_returns_an_empty_array(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/2/invoice-adjustments", nil)
	req.SetPathValue("clientId", "2")
	w := httptest.NewRecorder()

	ia := shared.InvoiceAdjustments{}

	mock := &mockService{invoiceAdjustments: ia}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	_ = server.getInvoiceAdjustments(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	expected := `[]`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(w.Body.String()))
	assert.Equal(t, 2, mock.expectedIds[0])
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getInvoiceAdjustments_error(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1/invoice-adjustments", nil)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	mock := &mockService{err: pgx.ErrTooManyRows}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	err := server.getInvoiceAdjustments(w, req)

	assert.Error(t, err)
}
