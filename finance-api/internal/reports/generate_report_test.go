package reports

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/db"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
	"time"
)

type MockFileStorage struct {
	versionId  string
	bucketName string
	filename   string
	file       io.Reader
	err        error
}

func (m *MockFileStorage) PutFile(ctx context.Context, bucketName string, fileName string, file io.Reader) (*string, error) {
	m.bucketName = bucketName
	m.filename = fileName
	m.file = file

	return &m.versionId, m.err
}

type MockNotify struct {
	payload notify.Payload
	err     error
}

func (m *MockNotify) Send(ctx context.Context, payload notify.Payload) error {
	m.payload = payload
	return m.err
}

type MockDb struct {
	query db.ReportQuery
	rows  [][]string
}

func (m *MockDb) Run(ctx context.Context, query db.ReportQuery) ([][]string, error) {
	m.query = query
	return m.rows, nil
}

func (m *MockDb) Close() {}

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
				ReportType: shared.ReportTypeAgedDebt,
				ToDate:     &toDate,
				FromDate:   &fromDate,
			},
			expectedQuery: &db.AgedDebt{FromDate: &fromDate, ToDate: &toDate},
		},
		{
			name: "Aged Debt By Customer",
			reportRequest: shared.ReportRequest{
				ReportType: shared.ReportTypeAgedDebtByCustomer,
			},
			expectedQuery: &db.AgedDebtByCustomer{},
		},
		{
			name: "Bad Debt Write Off",
			reportRequest: shared.ReportRequest{
				ReportType: shared.ReportTypeBadDebtWriteOff,
			},
			expectedQuery: &db.BadDebtWriteOff{},
		},
		{
			name: "Receipts",
			reportRequest: shared.ReportRequest{
				ReportType: shared.ReportTypeTotalReceipts,
				ToDate:     &toDate,
				FromDate:   &fromDate,
			},
			expectedQuery: &db.Receipts{FromDate: &fromDate, ToDate: &toDate},
		},
		{
			name: "Customer Credit",
			reportRequest: shared.ReportRequest{
				ReportType: shared.ReportTypeUnappliedReceipts,
			},
			expectedQuery: &db.CustomerCredit{},
		},
		{
			name: "Unknown",
			reportRequest: shared.ReportRequest{
				ReportType: shared.ReportAccountTypeUnknown,
			},
			expectedErr: fmt.Errorf("unknown query"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFileStorage := MockFileStorage{}
			mockNotify := MockNotify{}
			mockDb := MockDb{}

			client := NewClient(nil, &mockFileStorage, &mockNotify)
			client.db = &mockDb

			ctx := context.Background()
			timeNow, _ := time.Parse("2006-01-02", "2024-01-01")

			err := client.GenerateAndUploadReport(ctx, tt.reportRequest, timeNow)

			switch expected := tt.expectedQuery.(type) {
			case *db.AgedDebt:
				actual, ok := mockDb.query.(*db.AgedDebt)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Nil(t, err)
			case *db.AgedDebtByCustomer:
				actual, ok := mockDb.query.(*db.AgedDebtByCustomer)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Nil(t, err)
			case *db.BadDebtWriteOff:
				actual, ok := mockDb.query.(*db.BadDebtWriteOff)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Nil(t, err)
			case *db.Receipts:
				actual, ok := mockDb.query.(*db.Receipts)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Nil(t, err)
			case *db.CustomerCredit:
				actual, ok := mockDb.query.(*db.CustomerCredit)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Nil(t, err)
			default:
				assert.Equal(t, tt.expectedErr, err)
			}
			if err == nil {
				assert.NoError(t, os.Remove(mockFileStorage.filename))
			}
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
	_ = os.Setenv("FINANCE_ADMIN_PREFIX", "/finance")

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
