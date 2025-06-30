package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_updatePaymentMethod(t *testing.T) {
	var b bytes.Buffer

	_ = json.NewEncoder(&b).Encode(shared.UpdatePaymentMethod{PaymentMethod: shared.PaymentMethodDemanded})
	req := httptest.NewRequest(http.MethodPut, "/clients/1/payment-method", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	_ = server.updatePaymentMethod(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	expected := ""

	assert.Equal(t, expected, w.Body.String())
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestServer_updatePaymentMethod500Error(t *testing.T) {
	var b bytes.Buffer

	_ = json.NewEncoder(&b).Encode(shared.UpdatePaymentMethod{PaymentMethod: shared.PaymentMethodUnknown})
	req := httptest.NewRequest(http.MethodPut, "/clients/1/payment-method", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{err: errors.New("Something is wrong")}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.updatePaymentMethod(w, req)

	assert.Error(t, err)
}
