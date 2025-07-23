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

func TestServer_addDirectDebits(t *testing.T) {
	var b bytes.Buffer

	directDebitInfo := shared.AddDirectDebit{
		AccountName:   "Mrs Account Holder",
		SortCode:      "00-10-10",
		AccountNumber: "12345678",
	}
	_ = json.NewEncoder(&b).Encode(directDebitInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/direct-debits", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{addDirectDebit: directDebitInfo}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	_ = server.addDirectDebit(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	expected := ""

	assert.Equal(t, expected, w.Body.String())
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestServer_addDirectDebitValidationErrors(t *testing.T) {
	var b bytes.Buffer

	directDebitInfo := shared.AddDirectDebit{
		AccountName:   "",
		SortCode:      "",
		AccountNumber: "",
	}
	_ = json.NewEncoder(&b).Encode(directDebitInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/direct-debits", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{addDirectDebit: directDebitInfo}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.addDirectDebit(w, req)

	expected := apierror.ValidationError{Errors: apierror.ValidationErrors{
		"AccountName": {
			"required": "This field AccountName needs to be looked at required",
		},
		"AccountNumber": {
			"required": "This field AccountNumber needs to be looked at required",
		},
		"SortCode": {
			"required": "This field SortCode needs to be looked at required",
		},
	}}
	assert.Equal(t, expected, err)
}

func TestServer_addDirectDebit500Error(t *testing.T) {
	var b bytes.Buffer
	directDebitInfo := shared.AddDirectDebit{
		AccountName:   "",
		SortCode:      "",
		AccountNumber: "",
	}
	_ = json.NewEncoder(&b).Encode(directDebitInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/direct-debits", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{addDirectDebit: directDebitInfo, err: errors.New("something is wrong")}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.addDirectDebit(w, req)
	assert.Error(t, err)
}
