package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/db"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestGenerateAndUploadReport(t *testing.T) {
	toDate := shared.NewDate("2024-01-01")
	fromDate := shared.NewDate("2024-10-01")

	tests := []struct {
		name          string
		reportRequest shared.ReportRequest
		expectedQuery db.ReportQuery
		expectedErr   error
	}{
		{
			name: "Aged Debt",
			reportRequest: shared.ReportRequest{
				ReportType:        "AccountsReceivable",
				ReportAccountType: "AgedDebt",
				ToDateField:       &toDate,
				FromDateField:     &fromDate,
			},
			expectedQuery: &db.AgedDebt{FromDate: &fromDate, ToDate: &toDate},
		},
		{
			name: "Aged Debt By Customer",
			reportRequest: shared.ReportRequest{
				ReportType:        "AccountsReceivable",
				ReportAccountType: "AgedDebtByCustomer",
			},
			expectedQuery: &db.AgedDebtByCustomer{},
		},
		{
			name: "Unknown",
			reportRequest: shared.ReportRequest{
				ReportType:        "AccountsReceivable",
				ReportAccountType: "Garbleglarg",
			},
			expectedErr: fmt.Errorf("Unknown query"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReports := MockReports{}
			mockFileStorage := MockFileStorage{}
			mockHttpClient := MockHttpClient{}

			service := NewService()

			ctx := context.Background()
			timeNow, _ := time.Parse("2006-01-02", "2024-01-01")

			GetDoFunc = func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(bytes.NewReader([]byte{})),
				}, nil
			}

			err := service.GenerateAndUploadReport(ctx, tt.reportRequest, timeNow)

			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, &tt.expectedQuery, &mockReports.query)
		})
	}
}

func TestCreateDownloadNotifyPayload(t *testing.T) {
	emailAddress := "test@email.com"
	reportName := "test report"
	downloadRequest := shared.DownloadRequest{
		Key:       "test.csv",
		VersionId: "1",
	}
	uid, _ := downloadRequest.Encode()
	requestedDate, _ := time.Parse("2006-01-02 15:04:05", "2024-01-01 13:37:00")
	_ = os.Setenv("SIRIUS_PUBLIC_URL", "www.sirius.com")
	_ = os.Setenv("PREFIX", "/finance")

	want := notify.Payload{
		EmailAddress: emailAddress,
		TemplateId:   reportRequestedTemplateId,
		Personalisation: reportRequestedNotifyPersonalisation{
			FileLink:          fmt.Sprintf("www.sirius.com/finance/download?uid=%s", uid),
			ReportName:        reportName,
			RequestedDate:     "2024-01-01",
			RequestedDateTime: "2024-01-01 13:37:00",
		},
	}

	payload, err := createDownloadNotifyPayload(emailAddress, downloadRequest.Key, &downloadRequest.VersionId, requestedDate, reportName)

	assert.Equal(t, want, payload)
	assert.Nil(t, err)
}
