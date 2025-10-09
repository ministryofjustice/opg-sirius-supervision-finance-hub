package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
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
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	_ = server.updatePendingInvoiceAdjustment(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

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

	mock := &mockService{errs: map[string]error{"UpdatePendingInvoiceAdjustment": errors.New("something is wrong")}}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
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

	mock := &mockService{errs: map[string]error{"UpdatePendingInvoiceAdjustment": errors.New("something is wrong")}}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.updatePendingInvoiceAdjustment(w, req)

	expected := apierror.ValidationError{Errors: apierror.ValidationErrors{
		"Status": {
			"oneof": "This field Status needs to be looked at oneof",
		},
	}}
	assert.Equal(t, expected, err)
}
