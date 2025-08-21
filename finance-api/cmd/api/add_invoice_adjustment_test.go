package api

import (
	"bytes"
	"errors"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer_AddInvoiceAdjustment(t *testing.T) {
	validator, _ := validation.New()

	tests := []struct {
		name string
		body string
		err  error
	}{
		{
			name: "success",
			body: `{
				"adjustmentType": "CREDIT MEMO",
				"notes":"Some notes here", 
				"amount": 12345
			 }`,
			err: nil,
		},
		{
			name: "internal server error",
			body: `{
				"adjustmentType": "CREDIT MEMO",
				"notes":"Some notes here", 
				"amount": 12345
			 }`,
			err: errors.New("something is wrong"),
		},
		{
			name: "bad request",
			body: `{
				"adjustmentType": "CREDIT MEMO",
				"notes":"Some notes here", 
				"amount": 12345678
			 }`,
			err: apierror.BadRequestsError([]string{"Amount entered must be equal to or less than £420"}),
		},
		{
			name: "validation error",
			body: `{
				"adjustmentType": "UNKNOWN",
				"notes":"` + string(bytes.Repeat([]byte{byte('a')}, 1001)) + `", 
				"amount": -12345
			 }`,
			err: apierror.ValidationError{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/clients/1/invoices/2/invoice-adjustments", strings.NewReader(tt.body))
			req.SetPathValue("clientId", "1")
			req.SetPathValue("invoiceId", "2")
			w := httptest.NewRecorder()

			mock := &mockService{err: tt.err, expectedIds: []int{1, 2}}
			server := NewServer(mock, nil, nil, nil, nil, validator, nil)
			err := server.addInvoiceAdjustment(w, req)

			assert.Equal(t, 1, mock.expectedIds[0])
			assert.Equal(t, 2, mock.expectedIds[1])
			if tt.err != nil {
				assert.ErrorAs(t, err, &tt.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAddInvoiceAdjustmentRequest_validation(t *testing.T) {
	validator, _ := validation.New()

	tests := []struct {
		name     string
		arg      shared.AddInvoiceAdjustmentRequest
		expected map[string]string
	}{
		{
			name: "adjustment type and notes",
			arg: shared.AddInvoiceAdjustmentRequest{
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
			arg: shared.AddInvoiceAdjustmentRequest{
				AdjustmentType:  shared.AdjustmentTypeCreditMemo,
				AdjustmentNotes: "abc",
			},
			expected: map[string]string{
				"Amount": "required_if",
			},
		},
		{
			name: "missing amount valid when not required",
			arg: shared.AddInvoiceAdjustmentRequest{
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
