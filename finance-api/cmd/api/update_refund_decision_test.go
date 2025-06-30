package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_UpdateRefundDecision(t *testing.T) {
	var b bytes.Buffer

	_ = json.NewEncoder(&b).Encode(shared.UpdateRefundStatus{Status: shared.RefundStatusApproved})
	req := httptest.NewRequest(http.MethodPut, "/clients/1/refunds/2", &b)
	req.SetPathValue("clientId", "1")
	req.SetPathValue("refundId", "2")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	_ = server.updateRefundDecision(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	expected := ""

	assert.Equal(t, expected, w.Body.String())
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestServer_UpdateRefundDecision500Error(t *testing.T) {
	var b bytes.Buffer

	_ = json.NewEncoder(&b).Encode(shared.UpdateRefundStatus{Status: shared.RefundStatusApproved})
	req := httptest.NewRequest(http.MethodPut, "/clients/1/refunds/2", &b)
	req.SetPathValue("clientId", "1")
	req.SetPathValue("refundId", "2")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{err: errors.New("Something is wrong")}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.updateRefundDecision(w, req)

	assert.Error(t, err)
}

func TestServer_UpdateRefundDecisionValidationError(t *testing.T) {
	var b bytes.Buffer

	_ = json.NewEncoder(&b).Encode(shared.UpdateRefundStatus{Status: shared.RefundStatusFulfilled})
	req := httptest.NewRequest(http.MethodPut, "/clients/1/refunds/2", &b)
	req.SetPathValue("clientId", "1")
	req.SetPathValue("refundId", "2")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.updateRefundDecision(w, req)

	expected := apierror.ValidationError{Errors: apierror.ValidationErrors{
		"Status": {
			"oneof": "This field Status needs to be looked at oneof",
		},
	}}
	assert.Equal(t, expected, err)
}
