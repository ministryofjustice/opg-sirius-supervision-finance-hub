package api

import (
	"bytes"
	"errors"
	"github.com/opg-sirius-finance-hub/finance-api/internal/validation"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer_PostLedgerEntry(t *testing.T) {
	validator, _ := validation.New()

	tests := []struct {
		name   string
		body   string
		err    error
		status int
	}{
		{
			name: "success",
			body: `{
				"adjustmentType": "CREDIT MEMO",
				"notes":"Some notes here", 
				"amount": 12345
			 }`,
			err:    nil,
			status: 201,
		},
		{
			name: "internal server error",
			body: `{
				"adjustmentType": "CREDIT MEMO",
				"notes":"Some notes here", 
				"amount": 12345
			 }`,
			err:    errors.New("something is wrong"),
			status: 500,
		},
		{
			name: "bad request",
			body: `{
				"adjustmentType": "CREDIT MEMO",
				"notes":"Some notes here", 
				"amount": 12345678
			 }`,
			err:    shared.BadRequest{Field: "Amount", Reason: "Amount entered must be equal to or less than £420"},
			status: 400,
		},
		{
			name: "validation error",
			body: `{
				"adjustmentType": "UNKNOWN",
				"notes":"` + string(bytes.Repeat([]byte{byte('a')}, 1001)) + `", 
				"amount": -12345
			 }`,
			err:    nil,
			status: 422,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/clients/1/invoices/2/ledger-entries", strings.NewReader(tt.body))
			req.SetPathValue("clientId", "1")
			req.SetPathValue("invoiceId", "2")
			w := httptest.NewRecorder()

			mock := &mockService{err: tt.err, expectedIds: []int{1, 2}}
			server := Server{Service: mock, Validator: validator}
			server.PostLedgerEntry(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, 1, mock.expectedIds[0])
			assert.Equal(t, 2, mock.expectedIds[1])
			assert.Equal(t, tt.status, w.Code)
			if tt.err != nil {
				assert.Contains(t, w.Body.String(), tt.err.Error())
			}
		})
	}
}

func TestCreateLedgerEntryRequest_validation(t *testing.T) {
	validator, _ := validation.New()

	tests := []struct {
		name     string
		arg      shared.CreateLedgerEntryRequest
		expected map[string]string
	}{
		{
			name: "adjustment type and notes",
			arg: shared.CreateLedgerEntryRequest{
				AdjustmentType:  shared.AdjustmentTypeUnknown,
				AdjustmentNotes: string(bytes.Repeat([]byte{byte('a')}, 1001)),
			},
			expected: map[string]string{
				"AdjustmentType":  "valid-enum",
				"AdjustmentNotes": "thousand-character-limit",
			},
		},
		{
			name: "amount when required",
			arg: shared.CreateLedgerEntryRequest{
				AdjustmentType:  shared.AdjustmentTypeAddCredit,
				AdjustmentNotes: "abc",
			},
			expected: map[string]string{
				"Amount": "required_if",
			},
		},
		{
			name: "missing amount valid when not required",
			arg: shared.CreateLedgerEntryRequest{
				AdjustmentType:  shared.AdjustmentTypeWriteOff,
				AdjustmentNotes: "abc",
			},
			expected: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validator.ValidateStruct(tt.arg).Errors
			assert.Len(t, errs, len(tt.expected))
			for key, val := range errs {
				assert.NotEmpty(t, val[tt.expected[key]])
			}
		})
	}
}
