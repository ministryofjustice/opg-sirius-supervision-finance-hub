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

func TestServer_addRefund(t *testing.T) {
	var b bytes.Buffer

	refund := shared.BankDetails{
		Name:     "Mr Reginald Refund",
		Account:  "12345678",
		SortCode: "11-22-33",
	}
	_ = json.NewEncoder(&b).Encode(refund)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/refunds", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{refund: refund}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	_ = server.addRefund(w, req)

	res := w.Result()
	defer res.Body.Close()

	expected := ""

	assert.Equal(t, expected, w.Body.String())
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestServer_addRefundValidationErrors(t *testing.T) {
	var b bytes.Buffer

	refund := shared.BankDetails{
		Name:     "",
		Account:  "",
		SortCode: "",
	}
	_ = json.NewEncoder(&b).Encode(refund)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/refunds", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{refund: refund}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.addRefund(w, req)

	expected := apierror.ValidationError{Errors: apierror.ValidationErrors{
		"AccountName": {
			"required": "This field AccountName needs to be looked at required",
		},
		"Account": {
			"required": "This field Account needs to be looked at required",
		},
		"SortCode": {
			"required": "This field SortCode needs to be looked at required",
		},
	}}
	assert.Equal(t, expected, err)

}

func TestServer_addRefund500Error(t *testing.T) {
	var b bytes.Buffer
	refund := shared.BankDetails{
		Name:     "Mr Reginald Refund",
		Account:  "12345678",
		SortCode: "11-22-33",
	}
	_ = json.NewEncoder(&b).Encode(refund)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/refunds", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{refund: refund, err: errors.New("something is wrong")}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.addRefund(w, req)
	assert.Error(t, err)
}
