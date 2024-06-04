package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/opg-sirius-finance-hub/finance-api/internal/validation"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_updatePendingInvoiceAdjustment(t *testing.T) {
	var b bytes.Buffer

	_ = json.NewEncoder(&b).Encode(shared.UpdateInvoiceAdjustment{Status: "APPROVED"})
	req := httptest.NewRequest(http.MethodPut, "/clients/1/invoice-adjustments/1", &b)
	req.SetPathValue("id", "1")
	req.SetPathValue("ledgerId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{}
	server := Server{Service: mock, Validator: validator}
	server.updatePendingInvoiceAdjustment(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	expected := ""

	assert.Equal(t, expected, string(data))
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestServer_updatePendingInvoiceAdjustment500Error(t *testing.T) {
	var b bytes.Buffer

	_ = json.NewEncoder(&b).Encode(shared.UpdateInvoiceAdjustment{Status: "APPROVED"})
	req := httptest.NewRequest(http.MethodPut, "/clients/1/invoice-adjustments/1", &b)
	req.SetPathValue("id", "1")
	req.SetPathValue("ledgerId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{err: errors.New("Something is wrong")}
	server := Server{Service: mock, Validator: validator}
	server.updatePendingInvoiceAdjustment(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestServer_updatePendingInvoiceAdjustmentValidationError(t *testing.T) {
	var b bytes.Buffer

	_ = json.NewEncoder(&b).Encode(nil)
	req := httptest.NewRequest(http.MethodPut, "/clients/1/invoice-adjustments/1", &b)
	req.SetPathValue("id", "1")
	req.SetPathValue("ledgerId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{err: errors.New("Something is wrong")}
	server := Server{Service: mock, Validator: validator}
	server.updatePendingInvoiceAdjustment(w, req)

	res := w.Result()
	defer res.Body.Close()

	expectedError := shared.BadRequest{Field: "Status", Reason: "This field Status needs to be looked at oneof"}

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), expectedError.Error())
}
