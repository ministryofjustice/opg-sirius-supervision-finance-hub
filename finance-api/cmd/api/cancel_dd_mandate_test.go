package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func TestServer_cancelDirectDebitMandate(t *testing.T) {
	var b bytes.Buffer

	data := shared.CancelMandate{
		AllPayCustomer: shared.AllPayCustomer{
			Surname:         "Nameson",
			ClientReference: "1234567T",
		},
	}
	_ = json.NewEncoder(&b).Encode(data)
	req := httptest.NewRequest(http.MethodDelete, "/clients/1/direct-debit", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	_ = server.cancelDirectDebitMandate(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "CancelDirectDebitMandate", mock.called[0])
}

func TestServer_cancelDirectDebitMandate_500Error(t *testing.T) {
	var b bytes.Buffer

	data := shared.CancelMandate{
		AllPayCustomer: shared.AllPayCustomer{
			Surname:         "Nameson",
			ClientReference: "1234567T",
		},
	}
	_ = json.NewEncoder(&b).Encode(data)
	req := httptest.NewRequest(http.MethodDelete, "/clients/1/direct-debit", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{errs: map[string]error{"CancelDirectDebitMandate": errors.New("something is wrong")}}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.cancelDirectDebitMandate(w, req)

	assert.Error(t, err)
}
