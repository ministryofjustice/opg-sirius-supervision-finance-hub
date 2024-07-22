package api

import (
	"github.com/jackc/pgx/v5"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_getPermittedAdjustments(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/2/invoices/1/permitted-adjustments", nil)
	req.SetPathValue("invoiceId", "1")
	w := httptest.NewRecorder()

	types := []shared.AdjustmentType{shared.AdjustmentTypeCreditMemo, shared.AdjustmentTypeDebitMemo}

	mock := &mockService{adjustmentTypes: types}
	server := Server{Service: mock}
	server.getPermittedAdjustments(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	expected := `["CREDIT MEMO","DEBIT MEMO"]`

	assert.Equal(t, expected, string(data))
	assert.Equal(t, 1, mock.expectedIds[0])
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getPermittedAdjustments_invoiceNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/2/invoices/1/permitted-adjustments", nil)
	req.SetPathValue("invoiceId", "1")
	w := httptest.NewRecorder()

	mock := &mockService{err: pgx.ErrNoRows}
	server := Server{Service: mock}
	server.getPermittedAdjustments(w, req)

	res := w.Result()

	assert.Equal(t, 404, res.StatusCode)
}

func TestServer_getPermittedAdjustments_error(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/2/invoices/1/permitted-adjustments", nil)
	req.SetPathValue("invoiceId", "1")
	w := httptest.NewRecorder()

	mock := &mockService{err: pgx.ErrTooManyRows}
	server := Server{Service: mock}
	server.getPermittedAdjustments(w, req)

	res := w.Result()

	assert.Equal(t, 500, res.StatusCode)
}
