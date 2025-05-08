package reports

import (
	"bytes"
	"context"
	"encoding/csv"
	"github.com/ministryofjustice/opg-go-common/telemetry"
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
	data       io.Reader
	err        error
}

func (m *MockFileStorage) PutFile(ctx context.Context, bucketName string, fileName string, data *bytes.Buffer) (*string, error) {
	m.bucketName = bucketName
	m.filename = fileName
	m.data = data

	return &m.versionId, m.err
}

func (m *MockFileStorage) StreamFile(ctx context.Context, bucketName string, fileName string, stream io.ReadCloser) (*string, error) {
	m.bucketName = bucketName
	m.filename = fileName
	m.data = stream
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

func (m *MockDb) CopyStream(ctx context.Context, query db.ReportQuery) (io.ReadCloser, error) {
	m.query = query

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	for _, row := range m.rows {
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
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
		expectedTemplate string
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
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "Aged Debt By Customer",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeAgedDebtByCustomer),
			},
			expectedQuery:    &db.AgedDebtByCustomer{},
			expectedFilename: "AgedDebtByCustomer_01:01:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
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
			expectedTemplate: reportRequestedTemplateId,
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
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "Bad Debt Write Off",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeBadDebtWriteOff),
			},
			expectedQuery:    &db.BadDebtWriteOff{},
			expectedFilename: "BadDebtWriteOff_01:01:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
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
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "Customer Credit",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeUnappliedReceipts),
			},
			expectedQuery:    &db.CustomerCredit{},
			expectedFilename: "UnappliedReceipts_01:01:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "Fee Accrual",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeFeeAccrual),
			},
			expectedQuery:    nil,
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "NonReceiptTransactions",
			reportRequest: shared.ReportRequest{
				ReportType:      shared.ReportsTypeJournal,
				JournalType:     toPtr(shared.JournalTypeNonReceiptTransactions),
				TransactionDate: &toDate,
			},
			expectedQuery:    &db.NonReceiptTransactions{Date: &toDate},
			expectedFilename: "NonReceiptTransactions_01:01:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "ReceiptTransactions",
			reportRequest: shared.ReportRequest{
				ReportType:      shared.ReportsTypeJournal,
				JournalType:     toPtr(shared.JournalTypeReceiptTransactions),
				TransactionDate: &toDate,
			},
			expectedQuery:    &db.ReceiptTransactions{Date: &toDate},
			expectedFilename: "ReceiptTransactions_01:01:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "UnappliedTransactions",
			reportRequest: shared.ReportRequest{
				ReportType:      shared.ReportsTypeJournal,
				JournalType:     toPtr(shared.JournalTypeUnappliedTransactions),
				TransactionDate: &toDate,
			},
			expectedQuery:    &db.UnappliedTransactions{Date: &toDate},
			expectedFilename: "RefundUnappliedTransactions_01:01:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "Unknown",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeUnknown),
			},
			expectedTemplate: reportFailedTemplateId,
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
			expectedTemplate: reportRequestedTemplateId,
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
			expectedTemplate: reportRequestedTemplateId,
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
			expectedTemplate: reportRequestedTemplateId,
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
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "Unapplied payments",
			reportRequest: shared.ReportRequest{
				ReportType:      shared.ReportsTypeSchedule,
				ScheduleType:    toPtr(shared.ScheduleTypeUnappliedPayments),
				TransactionDate: &fromDate,
			},
			expectedQuery: &db.UnapplyReapplySchedule{
				Date:         &fromDate,
				ScheduleType: toPtr(shared.ScheduleTypeUnappliedPayments),
			},
			expectedFilename: "schedule_UnappliedPayments_10:10:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFileStorage := MockFileStorage{}
			mockNotify := MockNotify{}
			mockDb := MockDb{}

			client := NewClient(nil, &mockFileStorage, &mockNotify, &Envs{ReportsBucket: "test"})
			client.db = &mockDb

			ctx := telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test"))
			timeNow, _ := time.Parse("2006-01-02", "2024-01-01")

			client.GenerateAndUploadReport(ctx, tt.reportRequest, timeNow)

			switch expected := tt.expectedQuery.(type) {
			case *db.AgedDebt:
				actual, ok := mockDb.query.(*db.AgedDebt)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Equal(t, tt.expectedTemplate, mockNotify.payload.TemplateId)
			case *db.AgedDebtByCustomer:
				actual, ok := mockDb.query.(*db.AgedDebtByCustomer)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Equal(t, tt.expectedTemplate, mockNotify.payload.TemplateId)
			case *db.PaidInvoices:
				actual, ok := mockDb.query.(*db.PaidInvoices)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Equal(t, tt.expectedTemplate, mockNotify.payload.TemplateId)
			case *db.InvoiceAdjustments:
				actual, ok := mockDb.query.(*db.InvoiceAdjustments)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Equal(t, tt.expectedTemplate, mockNotify.payload.TemplateId)
			case *db.BadDebtWriteOff:
				actual, ok := mockDb.query.(*db.BadDebtWriteOff)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Equal(t, tt.expectedTemplate, mockNotify.payload.TemplateId)
			case *db.Receipts:
				actual, ok := mockDb.query.(*db.Receipts)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Equal(t, tt.expectedTemplate, mockNotify.payload.TemplateId)
			case *db.NonReceiptTransactions:
				actual, ok := mockDb.query.(*db.NonReceiptTransactions)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Equal(t, tt.expectedTemplate, mockNotify.payload.TemplateId)
			case *db.ReceiptTransactions:
				actual, ok := mockDb.query.(*db.ReceiptTransactions)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Equal(t, tt.expectedTemplate, mockNotify.payload.TemplateId)
			case *db.UnappliedTransactions:
				actual, ok := mockDb.query.(*db.UnappliedTransactions)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Equal(t, tt.expectedTemplate, mockNotify.payload.TemplateId)
			case *db.CustomerCredit:
				actual, ok := mockDb.query.(*db.CustomerCredit)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Equal(t, tt.expectedTemplate, mockNotify.payload.TemplateId)
			case *db.PaymentsSchedule:
				actual, ok := mockDb.query.(*db.PaymentsSchedule)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Equal(t, tt.expectedTemplate, mockNotify.payload.TemplateId)
			case *db.InvoicesSchedule:
				actual, ok := mockDb.query.(*db.InvoicesSchedule)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Equal(t, tt.expectedTemplate, mockNotify.payload.TemplateId)
			case *db.AdjustmentsSchedule:
				actual, ok := mockDb.query.(*db.AdjustmentsSchedule)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Equal(t, tt.expectedTemplate, mockNotify.payload.TemplateId)
			case *db.UnapplyReapplySchedule:
				actual, ok := mockDb.query.(*db.UnapplyReapplySchedule)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Equal(t, tt.expectedTemplate, mockNotify.payload.TemplateId)
			default:
				assert.Equal(t, tt.expectedTemplate, mockNotify.payload.TemplateId)
			}

			assert.Equal(t, tt.expectedFilename, mockFileStorage.filename)
			_ = os.Remove(mockFileStorage.filename)
		})
	}
}
func TestSendSuccessNotification(t *testing.T) {
	mockNotify := MockNotify{}
	client := &Client{notify: &mockNotify, envs: &Envs{FinanceAdminURL: "http://example.com"}}

	ctx := context.Background()
	emailAddress := "test@example.com"
	filename := "test_report.csv"
	versionId := "12345"
	requestedDate, _ := time.Parse("2006-01-02", "2024-01-01")
	reportName := "Test Report"

	err := client.sendSuccessNotification(ctx, emailAddress, filename, &versionId, requestedDate, reportName)
	assert.Nil(t, err)

	expectedPayload := notify.Payload{
		EmailAddress: emailAddress,
		TemplateId:   reportRequestedTemplateId,
		Personalisation: reportRequestedNotifyPersonalisation{
			FileLink:          "http://example.com/download?uid=eyJLZXkiOiJ0ZXN0X3JlcG9ydC5jc3YiLCJWZXJzaW9uSWQiOiIxMjM0NSJ9",
			ReportName:        reportName,
			RequestedDate:     "2024-01-01",
			RequestedDateTime: "2024-01-01 00:00:00",
		},
	}

	assert.Equal(t, expectedPayload, mockNotify.payload)
}

func TestSendFailureNotification(t *testing.T) {
	mockNotify := MockNotify{}
	client := &Client{notify: &mockNotify}

	ctx := context.Background()
	emailAddress := "test@example.com"
	requestedDate, _ := time.Parse("2006-01-02", "2024-01-01")
	reportName := "Test Report"

	err := client.sendFailureNotification(ctx, emailAddress, requestedDate, reportName)
	assert.Nil(t, err)

	expectedPayload := notify.Payload{
		EmailAddress: emailAddress,
		TemplateId:   reportFailedTemplateId,
		Personalisation: reportFailedNotifyPersonalisation{
			ReportName:        reportName,
			RequestedDate:     "2024-01-01",
			RequestedDateTime: "2024-01-01 00:00:00",
		},
	}

	assert.Equal(t, expectedPayload, mockNotify.payload)
}
