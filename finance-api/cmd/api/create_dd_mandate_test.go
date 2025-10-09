package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func TestServer_createDirectDebitMandateWithSchedule(t *testing.T) {
	var b bytes.Buffer

	data := shared.CreateMandate{
		AllPayCustomer: shared.AllPayCustomer{
			ClientReference: "11111111",
			Surname:         "Holder",
		},
		Address: shared.Address{
			Line1:    "1 Main Street",
			Town:     "Mainopolis",
			PostCode: "MP1 2PM",
		},
		BankAccount: struct {
			BankDetails shared.AllPayBankDetails `json:"bankDetails"`
		}{
			BankDetails: shared.AllPayBankDetails{
				AccountName:   "Mrs Account Holder",
				SortCode:      "30-33-30",
				AccountNumber: "12345678",
			},
		},
	}
	_ = json.NewEncoder(&b).Encode(data)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/direct-debit", &b)
	req.SetPathValue("clientId", "1")
	ctx := telemetry.ContextWithLogger(req.Context(), telemetry.NewLogger("test"))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{}

	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	_ = server.createDirectDebitMandate(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "CreateDirectDebitMandate", mock.called[0])
	assert.Equal(t, "CreateDirectDebitSchedule", mock.called[1])
}

func TestServer_createDirectDebitMandateWithoutSchedule(t *testing.T) {
	var b bytes.Buffer

	data := shared.CreateMandate{
		AllPayCustomer: shared.AllPayCustomer{
			ClientReference: "11111111",
			Surname:         "Holder",
		},
		Address: shared.Address{
			Line1:    "1 Main Street",
			Town:     "Mainopolis",
			PostCode: "MP1 2PM",
		},
		BankAccount: struct {
			BankDetails shared.AllPayBankDetails `json:"bankDetails"`
		}{
			BankDetails: shared.AllPayBankDetails{
				AccountName:   "Mrs Account Holder",
				SortCode:      "30-33-30",
				AccountNumber: "12345678",
			},
		},
	}
	_ = json.NewEncoder(&b).Encode(data)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/direct-debit", &b)
	req.SetPathValue("clientId", "1")
	ctx := telemetry.ContextWithLogger(req.Context(), telemetry.NewLogger("test"))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	_ = server.createDirectDebitMandate(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "CreateDirectDebitMandate", mock.called[0])
	assert.Equal(t, "CreateDirectDebitSchedule", mock.called[1])
}
