package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func Test_processUpload(t *testing.T) {
	notifyClient := &mockNotify{}
	service := &mockService{}

	server := NewServer(service, nil, nil, notifyClient, nil, nil, nil)

	tests := []struct {
		name   string
		upload shared.Upload
		err    error
	}{
		{
			name: "fail",
			upload: shared.Upload{
				UploadType:   shared.ReportTypeUploadUnknown,
				EmailAddress: "test@email.com",
				Base64Data:   "wrong",
			},
			err: apierror.BadRequest{},
		},
		{
			name: "pass",
			upload: shared.Upload{
				UploadType:   shared.ReportTypeUploadPaymentsMOTOCard,
				EmailAddress: "test@email.com",
				Base64Data:   base64.StdEncoding.EncodeToString([]byte("col1, col2\nabc,1")),
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		var body bytes.Buffer

		_ = json.NewEncoder(&body).Encode(tt.upload)
		r := httptest.NewRequest(http.MethodPost, "/upload", &body)
		ctx := auth.Context{
			Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test")),
			User:    &shared.User{ID: 1},
		}
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		err := server.processUpload(w, r)
		if tt.err != nil {
			assert.ErrorAs(t, err, &tt.err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func Test_processUploadFile(t *testing.T) {
	tests := []struct {
		name                string
		upload              Upload
		expectedPayload     notify.Payload
		expectedServiceCall string
	}{
		{
			name: "Unknown report",
			upload: Upload{
				UploadType:   shared.ReportTypeUploadUnknown,
				EmailAddress: "test@email.com",
				FileBytes:    bytes.NewReader([]byte("col1, col2\nabc,1")),
			},
			expectedPayload: notify.Payload{
				EmailAddress: "test@email.com",
				TemplateId:   notify.ProcessingErrorTemplateId,
				Personalisation: struct {
					Error      string `json:"error"`
					UploadType string `json:"upload_type"`
				}{
					"invalid upload type",
					"",
				},
			},
		},
		{
			name: "Unparseable csv",
			upload: Upload{
				UploadType:   shared.ReportTypeUploadUnknown,
				EmailAddress: "test@email.com",
				FileBytes:    bytes.NewReader([]byte("\"d,32")),
			},
			expectedPayload: notify.Payload{
				EmailAddress: "test@email.com",
				TemplateId:   notify.ProcessingErrorTemplateId,
				Personalisation: struct {
					Error      string `json:"error"`
					UploadType string `json:"upload_type"`
				}{
					"unable to read report",
					"",
				},
			},
		},
		{
			name: "Known payment upload",
			upload: Upload{
				UploadType:   shared.ReportTypeUploadPaymentsMOTOCard,
				EmailAddress: "test@email.com",
				FileBytes:    bytes.NewReader([]byte("col1, col2\nabc,1")),
			},
			expectedPayload: notify.Payload{
				EmailAddress: "test@email.com",
				TemplateId:   notify.ProcessingSuccessTemplateId,
				Personalisation: struct {
					UploadType string `json:"upload_type"`
				}{"Payments - MOTO card"},
			},
			expectedServiceCall: "ProcessPayments",
		},
		{
			name: "Known reversal upload",
			upload: Upload{
				UploadType:   shared.ReportTypeUploadMisappliedPayments,
				EmailAddress: "test@email.com",
				FileBytes:    bytes.NewReader([]byte("col1, col2\nabc,1")),
			},
			expectedPayload: notify.Payload{
				EmailAddress: "test@email.com",
				TemplateId:   notify.ProcessingSuccessTemplateId,
				Personalisation: struct {
					UploadType string `json:"upload_type"`
				}{"Payment Reversals - Misapplied payments"},
			},
			expectedServiceCall: "ProcessPaymentReversals",
		},
		{
			name: "Fulfilled refund",
			upload: Upload{
				UploadType:   shared.ReportTypeUploadFulfilledRefunds,
				EmailAddress: "test@email.com",
				FileBytes:    bytes.NewReader([]byte("col1, col2\nabc,1")),
			},
			expectedPayload: notify.Payload{
				EmailAddress: "test@email.com",
				TemplateId:   notify.ProcessingSuccessTemplateId,
				Personalisation: struct {
					UploadType string `json:"upload_type"`
				}{"Fulfilled refunds"},
			},
			expectedServiceCall: "ProcessFulfilledRefunds",
		},
		{
			name: "Known direct upload",
			upload: Upload{
				UploadType:   shared.ReportTypeUploadDebtChase,
				EmailAddress: "test@email.com",
				FileBytes:    bytes.NewReader([]byte("col1, col2\nabc,1")),
			},
			expectedPayload: notify.Payload{
				EmailAddress: "test@email.com",
				TemplateId:   notify.ProcessingSuccessTemplateId,
				Personalisation: struct {
					UploadType string `json:"upload_type"`
				}{"Debt chase"},
			},
			expectedServiceCall: "ProcessDirectUploadReport",
		},
	}
	for _, tt := range tests {
		notifyClient := &mockNotify{}
		service := &mockService{}

		server := NewServer(service, nil, nil, notifyClient, nil, nil, nil)

		ctx := auth.Context{
			Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test")),
			User:    &shared.User{ID: 1},
		}

		server.processUploadFile(ctx, tt.upload)
		assert.Equal(t, tt.expectedPayload, notifyClient.payload, tt.name)
		if tt.expectedServiceCall != "" {
			assert.Equal(t, tt.expectedServiceCall, service.called[0], tt.name)
		} else {
			assert.Len(t, service.called, 0, tt.name)
		}
	}
}

func Test_formatFailedLines(t *testing.T) {
	tests := []struct {
		name        string
		failedLines map[int]string
		want        []string
	}{
		{
			name:        "Empty",
			failedLines: map[int]string{},
			want:        []string(nil),
		},
		{
			name: "Unsorted lines",
			failedLines: map[int]string{
				5: "DATE_PARSE_ERROR",
				3: "CLIENT_NOT_FOUND",
				8: "DUPLICATE_PAYMENT",
				1: "DUPLICATE_PAYMENT",
			},
			want: []string{
				"Line 1: Duplicate payment line",
				"Line 3: Could not find a client with this court reference",
				"Line 5: Unable to parse date - please use the format DD/MM/YYYY",
				"Line 8: Duplicate payment line",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formattedLines := formatFailedLines(tt.failedLines)
			assert.Equal(t, tt.want, formattedLines)
		})
	}
}
