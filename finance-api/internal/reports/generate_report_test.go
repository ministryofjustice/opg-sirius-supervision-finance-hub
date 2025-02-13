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

func toPtr[T any](val T) *T {
	return &val
}

func TestGenerateAndUploadReport(t *testing.T) {
	toDate := shared.NewDate("2024-01-01")
	fromDate := shared.NewDate("2024-10-10")

	tests := []struct {
		name             string
		reportRequest    shared.ReportRequest
		expectedQuery    db.ReportQuery
		expectedFilename string
		expectedErr      error
	}{
		{
			name: "Aged Debt",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeAgedDebt),
				ToDate:                 &toDate,
				FromDate:               &fromDate,
			},
			expectedQuery:    &db.AgedDebt{FromDate: &fromDate, ToDate: &toDate},
			expectedFilename: "AgedDebt_01:01:2024.csv",
		},
		{
			name: "Aged Debt By Customer",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeAgedDebtByCustomer),
			},
			expectedQuery:    &db.AgedDebtByCustomer{},
			expectedFilename: "AgedDebtByCustomer_01:01:2024.csv",
		},
		{
			name: "Paid Invoices",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeARPaidInvoice),
				ToDate:                 &toDate,
				FromDate:               &fromDate,
			},
			expectedQuery:    &db.PaidInvoices{FromDate: &fromDate, ToDate: &toDate},
			expectedFilename: "ARPaidInvoice_01:01:2024.csv",
		},
		{
			name: "Invoice Adjustments",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeInvoiceAdjustments),
				ToDate:                 &toDate,
				FromDate:               &fromDate,
			},
			expectedQuery:    &db.InvoiceAdjustments{FromDate: &fromDate, ToDate: &toDate},
			expectedFilename: "InvoiceAdjustments_01:01:2024.csv",
		},
		{
			name: "Bad Debt Write Off",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeBadDebtWriteOff),
			},
			expectedQuery:    &db.BadDebtWriteOff{},
			expectedFilename: "BadDebtWriteOff_01:01:2024.csv",
		},
		{
			name: "Receipts",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeTotalReceipts),
				ToDate:                 &toDate,
				FromDate:               &fromDate,
			},
			expectedQuery:    &db.Receipts{FromDate: &fromDate, ToDate: &toDate},
			expectedFilename: "TotalReceipts_01:01:2024.csv",
		},
		{
			name: "Customer Credit",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeUnappliedReceipts),
			},
			expectedQuery:    &db.CustomerCredit{},
			expectedFilename: "UnappliedReceipts_01:01:2024.csv",
		},
		{
			name: "Unknown",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeUnknown),
			},
			expectedErr: fmt.Errorf("unimplemented accounts receivable query: %s", shared.AccountsReceivableTypeUnknown.Key()),
		},
		{
			name: "Online card payment schedule",
			reportRequest: shared.ReportRequest{
				ReportType:      shared.ReportsTypeSchedule,
				ScheduleType:    toPtr(shared.ScheduleTypeOnlineCardPayments),
				TransactionDate: &toDate,
			},
			expectedQuery: &db.PaymentsSchedule{
				Date:         &toDate,
				ScheduleType: toPtr(shared.ScheduleTypeOnlineCardPayments),
			},
			expectedFilename: "schedule_OnlineCardPayments_01:01:2024.csv",
		},
		{
			name: "SE Invoice schedule",
			reportRequest: shared.ReportRequest{
				ReportType:      shared.ReportsTypeSchedule,
				ScheduleType:    toPtr(shared.ScheduleTypeSEFeeInvoicesGeneral),
				TransactionDate: &toDate,
			},
			expectedQuery: &db.InvoicesSchedule{
				Date:         &toDate,
				ScheduleType: toPtr(shared.ScheduleTypeSEFeeInvoicesGeneral),
			},
			expectedFilename: "schedule_SEFeeInvoicesGeneral_01:01:2024.csv",
		},
		{
			name: "AD Fee Reduction adjustments schedule",
			reportRequest: shared.ReportRequest{
				ReportType:      shared.ReportsTypeSchedule,
				ScheduleType:    toPtr(shared.ScheduleTypeADFeeReductions),
				TransactionDate: &toDate,
			},
			expectedQuery: &db.AdjustmentsSchedule{
				Date:         &toDate,
				ScheduleType: toPtr(shared.ScheduleTypeADFeeReductions),
			},
			expectedFilename: "schedule_ADFeeReductions_01:01:2024.csv",
		},
		{
			name: "Minimal debit adjustments schedule",
			reportRequest: shared.ReportRequest{
				ReportType:      shared.ReportsTypeSchedule,
				ScheduleType:    toPtr(shared.ScheduleTypeMinimalManualDebits),
				TransactionDate: &fromDate,
			},
			expectedQuery: &db.AdjustmentsSchedule{
				Date:         &fromDate,
				ScheduleType: toPtr(shared.ScheduleTypeMinimalManualDebits),
			},
			expectedFilename: "schedule_MinimalManualDebits_10:10:2024.csv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFileStorage := MockFileStorage{}
			mockNotify := MockNotify{}
			mockDb := MockDb{}

			client := NewClient(nil, &mockFileStorage, &mockNotify, &Envs{ReportsBucket: "test"})
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
			case *db.PaidInvoices:
				actual, ok := mockDb.query.(*db.PaidInvoices)
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
			case *db.PaymentsSchedule:
				actual, ok := mockDb.query.(*db.PaymentsSchedule)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Nil(t, err)
			case *db.AdjustmentsSchedule:
				actual, ok := mockDb.query.(*db.AdjustmentsSchedule)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Nil(t, err)
			default:
				assert.Equal(t, tt.expectedErr, err)
			}

			assert.Equal(t, tt.expectedFilename, mockFileStorage.filename)
			_ = os.Remove(mockFileStorage.filename)
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

	client := NewClient(nil, nil, nil, &Envs{
		ReportsBucket:   "test",
		FinanceAdminURL: "www.sirius.com/finance",
	})
	payload, err := client.createDownloadNotifyPayload(emailAddress, downloadRequest.Key, &downloadRequest.VersionId, requestedDate, reportName)

	assert.Equal(t, want, payload)
	assert.Nil(t, err)
}
