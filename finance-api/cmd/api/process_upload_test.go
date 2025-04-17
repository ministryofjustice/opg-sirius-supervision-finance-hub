package api

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

func Test_processFinanceAdminUpload(t *testing.T) {
	ctx := auth.Context{
		Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test")),
		User:    &shared.User{ID: 1},
	}

	fileStorage := &mockFileStorage{}
	fileStorage.data = io.NopCloser(strings.NewReader("test"))
	notifyClient := &mockNotify{}
	service := &mockService{}

	server := NewServer(service, nil, fileStorage, notifyClient, nil, nil, nil)

	tests := []struct {
		name                string
		uploadType          shared.ReportUploadType
		fileStorageErr      error
		expectedPayload     notify.Payload
		expectedServiceCall string
	}{
		{
			name:       "Unknown report",
			uploadType: shared.ReportTypeUploadUnknown,
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
			name:           "S3 error",
			uploadType:     shared.ReportTypeUploadPaymentsOPGBACS,
			fileStorageErr: fmt.Errorf("test"),
			expectedPayload: notify.Payload{
				EmailAddress: "test@email.com",
				TemplateId:   notify.ProcessingErrorTemplateId,
				Personalisation: struct {
					Error      string `json:"error"`
					UploadType string `json:"upload_type"`
				}{
					"unable to download report",
					"Payments - OPG BACS",
				},
			},
		},
		{
			name:       "Known payment upload",
			uploadType: shared.ReportTypeUploadPaymentsMOTOCard,
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
			name:       "Known reversal upload",
			uploadType: shared.ReportTypeUploadMisappliedPayments,
			expectedPayload: notify.Payload{
				EmailAddress: "test@email.com",
				TemplateId:   notify.ProcessingSuccessTemplateId,
				Personalisation: struct {
					UploadType string `json:"upload_type"`
				}{"Payment Reversals - Misapplied payments"},
			},
			expectedServiceCall: "ProcessReversals",
		},
	}
	for _, tt := range tests {
		filename := "test.csv"
		emailAddress := "test@email.com"
		fileStorage.err = tt.fileStorageErr

		err := server.processUpload(ctx, shared.FinanceAdminUploadEvent{
			EmailAddress: emailAddress, Filename: filename, UploadType: tt.uploadType,
		})
		assert.Nil(t, err)
		assert.Equal(t, tt.expectedPayload, notifyClient.payload)
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
