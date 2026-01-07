package reports

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"os"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/db"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
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
	timeNow, _ := time.Parse("2006-01-02", "2024-02-02")

	tests := []struct {
		name             string
		reportRequest    shared.ReportRequest
		expectedQuery    db.ReportQuery
		expectedFilename string
		expectedTemplate string
	}{
		{
			name: "Aged Debt to date",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeAgedDebt),
				ToDate:                 &toDate,
			},
			expectedQuery: &db.AgedDebt{
				AgedDebtInput: db.AgedDebtInput{
					ToDate: &toDate,
					Today:  time.Now(),
				},
				ReportQuery: db.NewReportQuery(db.AgedDebtQuery)},
			expectedFilename: "AgedDebt_01:01:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "Aged Debt no date",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeAgedDebt),
			},
			expectedQuery: &db.AgedDebt{
				AgedDebtInput: db.AgedDebtInput{
					Today: time.Now(),
				},
				ReportQuery: db.NewReportQuery(db.AgedDebtQuery)},
			expectedFilename: "AgedDebt_02:02:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "Aged Debt By Customer",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeAgedDebtByCustomer),
			},
			expectedQuery: &db.AgedDebtByCustomer{
				AgedDebtByCustomerInput: db.AgedDebtByCustomerInput{Today: time.Now()},
				ReportQuery:             db.NewReportQuery(db.AgedDebtByCustomerQuery),
			},
			expectedFilename: "AgedDebtByCustomer_02:02:2024.csv",
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
			expectedQuery: &db.PaidInvoices{
				PaidInvoicesInput: db.PaidInvoicesInput{
					FromDate: &fromDate,
					ToDate:   &toDate,
				},
				ReportQuery: db.NewReportQuery(db.PaidInvoicesQuery)},
			expectedFilename: "ARPaidInvoice_02:02:2024.csv",
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
			expectedQuery: &db.InvoiceAdjustments{
				InvoiceAdjustmentsInput: db.InvoiceAdjustmentsInput{
					FromDate: &fromDate,
					ToDate:   &toDate,
				},
				ReportQuery: db.NewReportQuery(db.InvoiceAdjustmentsQuery)},
			expectedFilename: "InvoiceAdjustments_02:02:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "Bad Debt Write Off",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeBadDebtWriteOff),
			},
			expectedQuery:    &db.BadDebtWriteOff{ReportQuery: db.NewReportQuery(db.BadDebtWriteOffQuery)},
			expectedFilename: "BadDebtWriteOff_02:02:2024.csv",
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
			expectedQuery: &db.Receipts{
				ReceiptsInput: db.ReceiptsInput{
					FromDate: &fromDate,
					ToDate:   &toDate,
				},
				ReportQuery: db.NewReportQuery(db.ReceiptsQuery)},
			expectedFilename: "TotalReceipts_02:02:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "Customer Credit",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeUnappliedReceipts),
			},
			expectedQuery:    &db.CustomerCredit{ReportQuery: db.NewReportQuery(db.CustomerCreditQuery)},
			expectedFilename: "UnappliedReceipts_02:02:2024.csv",
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
			name: "Unknown accounts receivable",
			reportRequest: shared.ReportRequest{
				ReportType:             shared.ReportsTypeAccountsReceivable,
				AccountsReceivableType: toPtr(shared.AccountsReceivableTypeUnknown),
			},
			expectedTemplate: reportFailedTemplateId,
		},
		{
			name: "NonReceiptTransactions",
			reportRequest: shared.ReportRequest{
				ReportType:      shared.ReportsTypeJournal,
				JournalType:     toPtr(shared.JournalTypeNonReceiptTransactions),
				TransactionDate: &toDate,
			},
			expectedQuery: &db.NonReceiptTransactions{
				NonReceiptTransactionsInput: db.NonReceiptTransactionsInput{
					Date: &toDate,
				},
				ReportQuery: db.NewReportQuery(db.NonReceiptTransactionsQuery)},
			expectedFilename: "NonReceiptTransactions_01:01:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "NonReceiptTransactionsHistoric",
			reportRequest: shared.ReportRequest{
				ReportType:      shared.ReportsTypeJournal,
				JournalType:     toPtr(shared.JournalTypeNonReceiptTransactionsHistoric),
				TransactionDate: &toDate,
			},
			expectedQuery: &db.NonReceiptTransactionsHistoric{
				NonReceiptTransactionsHistoricInput: db.NonReceiptTransactionsHistoricInput{
					Date: &toDate,
				},
				ReportQuery: db.NewReportQuery(db.NonReceiptTransactionsHistoricQuery)},
			expectedFilename: "NonReceiptTransactionsHistoric_01:01:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "ReceiptTransactionsHistoric",
			reportRequest: shared.ReportRequest{
				ReportType:      shared.ReportsTypeJournal,
				JournalType:     toPtr(shared.JournalTypeReceiptTransactionsHistoric),
				TransactionDate: &toDate,
			},
			expectedQuery: &db.ReceiptTransactionsHistoric{
				ReceiptTransactionsHistoricInput: db.ReceiptTransactionsHistoricInput{
					Date: &toDate,
				},
				ReportQuery: db.NewReportQuery(db.ReceiptTransactionsHistoricQuery)},
			expectedFilename: "ReceiptTransactionsHistoric_01:01:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "ReceiptTransactions",
			reportRequest: shared.ReportRequest{
				ReportType:      shared.ReportsTypeJournal,
				JournalType:     toPtr(shared.JournalTypeReceiptTransactions),
				TransactionDate: &toDate,
			},
			expectedQuery: &db.ReceiptTransactions{
				ReceiptTransactionsInput: db.ReceiptTransactionsInput{
					Date: &toDate,
				},
				ReportQuery: db.NewReportQuery(db.ReceiptTransactionsQuery)},
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
			expectedQuery: &db.UnappliedTransactions{
				UnappliedTransactionsInput: db.UnappliedTransactionsInput{
					Date: &toDate,
				},
				ReportQuery: db.NewReportQuery(db.UnappliedTransactionsQuery)},
			expectedFilename: "RefundUnappliedTransactions_01:01:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "Unknown journal",
			reportRequest: shared.ReportRequest{
				ReportType:      shared.ReportsTypeJournal,
				JournalType:     toPtr(shared.JournalTypeUnknown),
				TransactionDate: &toDate,
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
				PaymentsScheduleInput: db.PaymentsScheduleInput{
					Date:         &toDate,
					ScheduleType: toPtr(shared.ScheduleTypeOnlineCardPayments),
				},
				ReportQuery: db.NewReportQuery(db.PaymentsScheduleQuery),
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
				InvoicesScheduleInput: db.InvoicesScheduleInput{
					Date:         &toDate,
					ScheduleType: toPtr(shared.ScheduleTypeSEFeeInvoicesGeneral),
				},
				ReportQuery: db.NewReportQuery(db.InvoicesScheduleQuery),
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
				AdjustmentsScheduleInput: db.AdjustmentsScheduleInput{
					Date:         &toDate,
					ScheduleType: toPtr(shared.ScheduleTypeADFeeReductions),
				},
				ReportQuery: db.NewReportQuery(db.AdjustmentsScheduleQuery),
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
				AdjustmentsScheduleInput: db.AdjustmentsScheduleInput{
					Date:         &fromDate,
					ScheduleType: toPtr(shared.ScheduleTypeMinimalManualDebits),
				},
				ReportQuery: db.NewReportQuery(db.AdjustmentsScheduleQuery),
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
				UnapplyReapplyScheduleInput: db.UnapplyReapplyScheduleInput{
					Date:         &fromDate,
					ScheduleType: toPtr(shared.ScheduleTypeUnappliedPayments),
				},
				ReportQuery: db.NewReportQuery(db.UnapplyReapplyScheduleQuery),
			},
			expectedFilename: "schedule_UnappliedPayments_10:10:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "Refunds",
			reportRequest: shared.ReportRequest{
				ReportType:      shared.ReportsTypeSchedule,
				ScheduleType:    toPtr(shared.ScheduleTypeRefunds),
				TransactionDate: &fromDate,
			},
			expectedQuery: &db.RefundsSchedule{
				RefundsScheduleInput: db.RefundsScheduleInput{
					Date: &fromDate,
				},
				ReportQuery: db.NewReportQuery(db.RefundsScheduleQuery),
			},
			expectedFilename: "schedule_Refunds_10:10:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "Unknown schedule",
			reportRequest: shared.ReportRequest{
				ReportType:      shared.ReportsTypeSchedule,
				ScheduleType:    toPtr(shared.ScheduleTypeUnknown),
				TransactionDate: &toDate,
			},
			expectedTemplate: reportFailedTemplateId,
		},
		{
			name: "Debt chase",
			reportRequest: shared.ReportRequest{
				ReportType: shared.ReportsTypeDebt,
				DebtType:   toPtr(shared.DebtTypeFeeChase),
			},
			expectedQuery:    &db.FeeChase{ReportQuery: db.NewReportQuery(db.FeeChaseQuery)},
			expectedFilename: "debt_FeeChase_02:02:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "Final Fee Debt",
			reportRequest: shared.ReportRequest{
				ReportType: shared.ReportsTypeDebt,
				DebtType:   toPtr(shared.DebtTypeFinalFee),
			},
			expectedQuery:    &db.FinalFeeDebt{ReportQuery: db.NewReportQuery(db.FinalFeeDebtQuery)},
			expectedFilename: "debt_FinalFee_02:02:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "Approved refunds",
			reportRequest: shared.ReportRequest{
				ReportType: shared.ReportsTypeDebt,
				DebtType:   toPtr(shared.DebtTypeApprovedRefunds),
			},
			expectedQuery:    &db.ApprovedRefunds{},
			expectedFilename: "debt_ApprovedRefunds_02:02:2024.csv",
			expectedTemplate: reportRequestedTemplateId,
		},
		{
			name: "Unknown debt",
			reportRequest: shared.ReportRequest{
				ReportType: shared.ReportsTypeDebt,
				DebtType:   toPtr(shared.DebtTypeUnknown),
			},
			expectedTemplate: reportFailedTemplateId,
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

			client.GenerateAndUploadReport(ctx, tt.reportRequest, timeNow)

			switch expected := tt.expectedQuery.(type) {
			case *db.AgedDebt:
				actual, ok := mockDb.query.(*db.AgedDebt)
				//ignore milliseconds of difference in the queries for sanity
				actual.Today = actual.Today.Truncate(time.Second)
				expected.Today = expected.Today.Truncate(time.Second)
				assert.True(t, ok)
				assert.Equal(t, expected, actual)
				assert.Equal(t, tt.expectedTemplate, mockNotify.payload.TemplateId)
			case *db.AgedDebtByCustomer:
				actual, ok := mockDb.query.(*db.AgedDebtByCustomer)
				//ignore milliseconds of difference in the queries for sanity
				actual.Today = actual.Today.Truncate(time.Second)
				expected.Today = expected.Today.Truncate(time.Second)
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
