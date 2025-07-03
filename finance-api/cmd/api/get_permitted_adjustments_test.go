package api

import (
	"github.com/jackc/pgx/v5"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer_getPermittedAdjustments(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/2/invoices/1/permitted-adjustments", nil)
	req.SetPathValue("invoiceId", "1")
	w := httptest.NewRecorder()

	types := []shared.AdjustmentType{shared.AdjustmentTypeCreditMemo, shared.AdjustmentTypeDebitMemo}

	mock := &mockService{adjustmentTypes: types}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	_ = server.getPermittedAdjustments(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	expected := `["CREDIT MEMO","DEBIT MEMO"]`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(w.Body.String()))
	assert.Equal(t, 1, mock.expectedIds[0])
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getPermittedAdjustments_invoiceNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/2/invoices/1/permitted-adjustments", nil)
	req.SetPathValue("invoiceId", "1")
	w := httptest.NewRecorder()

	mock := &mockService{err: pgx.ErrNoRows}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	err := server.getPermittedAdjustments(w, req)

	expected := apierror.NotFoundError(pgx.ErrNoRows)
	assert.ErrorAs(t, err, &expected)
}

func TestServer_getPermittedAdjustments_error(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/2/invoices/1/permitted-adjustments", nil)
	req.SetPathValue("invoiceId", "1")
	w := httptest.NewRecorder()

	mock := &mockService{err: pgx.ErrTooManyRows}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	err := server.getPermittedAdjustments(w, req)

	assert.Error(t, err)
}
