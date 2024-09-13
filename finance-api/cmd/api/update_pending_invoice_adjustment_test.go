package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/opg-sirius-finance-hub/finance-api/internal/validation"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_updatePendingInvoiceAdjustment(t *testing.T) {
	var b bytes.Buffer

	_ = json.NewEncoder(&b).Encode(shared.UpdateInvoiceAdjustment{Status: shared.AdjustmentStatusApproved})
	req := httptest.NewRequest(http.MethodPut, "/clients/1/invoice-adjustments/2", &b)
	req.SetPathValue("clientId", "1")
	req.SetPathValue("adjustmentId", "2")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{}
	server := Server{Service: mock, Validator: validator}
	_ = server.updatePendingInvoiceAdjustment(w, req)

	res := w.Result()
	defer res.Body.Close()

	expected := ""

	assert.Equal(t, expected, w.Body.String())
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestServer_updatePendingInvoiceAdjustment500Error(t *testing.T) {
	var b bytes.Buffer

	_ = json.NewEncoder(&b).Encode(shared.UpdateInvoiceAdjustment{Status: shared.AdjustmentStatusApproved})
	req := httptest.NewRequest(http.MethodPut, "/clients/1/invoice-adjustments/2", &b)
	req.SetPathValue("clientId", "1")
	req.SetPathValue("adjustmentId", "2")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{err: errors.New("Something is wrong")}
	server := Server{Service: mock, Validator: validator}
	err := server.updatePendingInvoiceAdjustment(w, req)

	assert.Error(t, err)
}

func TestServer_updatePendingInvoiceAdjustmentValidationError(t *testing.T) {
	var b bytes.Buffer

	_ = json.NewEncoder(&b).Encode(shared.UpdateInvoiceAdjustment{Status: shared.AdjustmentStatusPending})
	req := httptest.NewRequest(http.MethodPut, "/clients/1/invoice-adjustments/2", &b)
	req.SetPathValue("clientId", "1")
	req.SetPathValue("adjustmentId", "2")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{err: errors.New("Something is wrong")}
	server := Server{Service: mock, Validator: validator}
	_ = server.updatePendingInvoiceAdjustment(w, req)

	res := w.Result()
	defer res.Body.Close()

	expectedError := shared.BadRequest{Field: "Status", Reason: "This field Status needs to be looked at oneof"}

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), expectedError.Error())
}
