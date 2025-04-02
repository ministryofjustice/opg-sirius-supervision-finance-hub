package service

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"
)

type createdLedgerAllocation struct {
	ledgerAmount     int
	ledgerType       string
	ledgerStatus     string
	datetime         time.Time
	allocationAmount int
	allocationStatus string
	invoiceId        int
}

type mockNotify struct {
	payload notify.Payload
	err     error
}

func (n *mockNotify) Send(ctx context.Context, payload notify.Payload) error {
	n.payload = payload
	return n.err
}

func (suite *IntegrationSuite) Test_processFinanceAdminUpload() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	fileStorage := &mockFileStorage{}
	fileStorage.file = io.NopCloser(strings.NewReader("test"))
	notifyClient := &mockNotify{}

	s := NewService(seeder.Conn, nil, fileStorage, notifyClient, nil)

	tests := []struct {
		name            string
		uploadType      string
		fileStorageErr  error
		expectedPayload notify.Payload
	}{
		{
			name:       "Unknown report",
			uploadType: "test",
			expectedPayload: notify.Payload{
				EmailAddress: "test@email.com",
				TemplateId:   notify.ProcessingErrorTemplateId,
				Personalisation: struct {
					Error      string `json:"error"`
					UploadType string `json:"upload_type"`
				}{
					"unknown upload type",
					"",
				},
			},
		},
		{
			name:           "S3 error",
			uploadType:     "test",
			fileStorageErr: fmt.Errorf("test"),
			expectedPayload: notify.Payload{
				EmailAddress: "test@email.com",
				TemplateId:   notify.ProcessingErrorTemplateId,
				Personalisation: struct {
					Error      string `json:"error"`
					UploadType string `json:"upload_type"`
				}{
					"unable to download report",
					"",
				},
			},
		},
		{
			name:       "Known report",
			uploadType: "PAYMENTS_MOTO_CARD",
			expectedPayload: notify.Payload{
				EmailAddress: "test@email.com",
				TemplateId:   notify.ProcessingSuccessTemplateId,
				Personalisation: struct {
					UploadType string `json:"upload_type"`
				}{"Payments - MOTO card"},
			},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			filename := "test.csv"
			emailAddress := "test@email.com"
			fileStorage.err = tt.fileStorageErr

			err := s.ProcessFinanceAdminUpload(suite.ctx, shared.FinanceAdminUploadEvent{
				EmailAddress: emailAddress, Filename: filename, UploadType: tt.uploadType,
			})
			assert.Nil(t, err)
			assert.Equal(t, tt.expectedPayload, notifyClient.payload)
		})
	}
}

func (suite *IntegrationSuite) Test_processPayments() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, 'invoice-1', 'DEMANDED', NULL, '1234');",
		"INSERT INTO finance_client VALUES (2, 2, 'invoice-2', 'DEMANDED', NULL, '12345');",
		"INSERT INTO finance_client VALUES (3, 3, 'invoice-3', 'DEMANDED', NULL, '123456');",
		"INSERT INTO invoice VALUES (1, 1, 1, 'AD', 'AD11223/19', '2023-04-01', '2025-03-31', 15000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO invoice VALUES (2, 2, 2, 'AD', 'AD11224/19', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO invoice VALUES (3, 3, 3, 'AD', 'AD11225/19', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO invoice VALUES (4, 3, 3, 'AD', 'AD11226/19', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
	)

	dispatch := &mockDispatch{}
	s := NewService(seeder.Conn, dispatch, nil, nil, nil)

	tests := []struct {
		name                      string
		records                   [][]string
		bankDate                  shared.Date
		expectedClientId          int
		expectedLedgerAllocations []createdLedgerAllocation
		want                      error
	}{
		{
			name: "Underpayment",
			records: [][]string{
				{"Ordercode", "BankDate", "Amount"},
				{
					"1234-1",
					"01/01/2024",
					"100",
				},
			},
			bankDate:         shared.NewDate("2024-01-17"),
			expectedClientId: 1,
			expectedLedgerAllocations: []createdLedgerAllocation{
				{
					10000,
					"MOTO CARD PAYMENT",
					"CONFIRMED",
					time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					10000,
					"ALLOCATED",
					1,
				},
			},
		},
		{
			name: "Overpayment",
			records: [][]string{
				{"Ordercode", "BankDate", "Amount"},
				{
					"12345",
					"01/01/2024",
					"250.1",
				},
			},
			bankDate:         shared.NewDate("2024-01-17"),
			expectedClientId: 2,
			expectedLedgerAllocations: []createdLedgerAllocation{
				{
					25010,
					"MOTO CARD PAYMENT",
					"CONFIRMED",
					time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					10000,
					"ALLOCATED",
					2,
				},
				{
					25010,
					"MOTO CARD PAYMENT",
					"CONFIRMED",
					time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					-15010,
					"UNAPPLIED",
					0,
				},
			},
		},
		{
			name: "Underpayment with multiple invoices",
			records: [][]string{
				{"Ordercode", "BankDate", "Amount"},
				{
					"123456",
					"01/01/2024",
					"50",
				},
			},
			bankDate:         shared.NewDate("2024-01-17"),
			expectedClientId: 3,
			expectedLedgerAllocations: []createdLedgerAllocation{
				{
					5000,
					"MOTO CARD PAYMENT",
					"CONFIRMED",
					time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					5000,
					"ALLOCATED",
					3,
				},
			},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			var failedLines map[int]string
			failedLines, err := s.processPayments(suite.ctx, tt.records, "PAYMENTS_MOTO_CARD", tt.bankDate)
			assert.Equal(t, tt.want, err)
			assert.Equal(t, map[int]string{}, failedLines)

			var createdLedgerAllocations []createdLedgerAllocation

			rows, _ := seeder.Query(suite.ctx,
				`SELECT l.amount, l.type, l.status, l.datetime, la.amount, la.status, la.invoice_id
						FROM ledger l
						LEFT JOIN ledger_allocation la ON l.id = la.ledger_id
					WHERE l.finance_client_id = $1`, tt.expectedClientId)

			for rows.Next() {
				var r createdLedgerAllocation
				_ = rows.Scan(&r.ledgerAmount, &r.ledgerType, &r.ledgerStatus, &r.datetime, &r.allocationAmount, &r.allocationStatus, &r.invoiceId)
				createdLedgerAllocations = append(createdLedgerAllocations, r)
			}

			assert.Equal(t, tt.expectedLedgerAllocations, createdLedgerAllocations)
		})
	}
}

func (suite *IntegrationSuite) Test_getPaymentDetails() {
	tests := []struct {
		name                   string
		record                 []string
		uploadType             string
		bankDate               shared.Date
		ledgerType             string
		index                  int
		expectedPaymentDetails shared.PaymentDetails
		expectedFailedLines    map[int]string
	}{
		{
			name:       "Moto Card Payment Record",
			record:     []string{"12345678", "01/01/2025", "200.12"},
			uploadType: shared.ReportTypeUploadPaymentsMOTOCard.Key(),
			bankDate:   shared.NewDate("2025-01-10"),
			ledgerType: shared.TransactionTypeMotoCardPayment.Key(),
			expectedPaymentDetails: shared.PaymentDetails{
				Amount:       20012,
				BankDate:     time.Date(2025, time.January, 10, 0, 0, 0, 0, time.UTC),
				CourtRef:     "12345678",
				LedgerType:   shared.TransactionTypeMotoCardPayment.Key(),
				ReceivedDate: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name:       "Online Card Payment Record",
			record:     []string{"87654321-abcdefg-garbleddata", "01/02/2025", "20.1"},
			uploadType: shared.ReportTypeUploadPaymentsOnlineCard.Key(),
			bankDate:   shared.NewDate("2025-02-10"),
			ledgerType: shared.TransactionTypeOnlineCardPayment.Key(),
			expectedPaymentDetails: shared.PaymentDetails{
				Amount:       2010,
				BankDate:     time.Date(2025, time.February, 10, 0, 0, 0, 0, time.UTC),
				CourtRef:     "87654321",
				LedgerType:   shared.TransactionTypeOnlineCardPayment.Key(),
				ReceivedDate: time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name:       "Supervision BACS Record",
			record:     []string{"", "", "", "", "01/02/2025", "", "550", "", "", "", "54326543"},
			uploadType: shared.ReportTypeUploadPaymentsSupervisionBACS.Key(),
			bankDate:   shared.NewDate("2025-02-10"),
			ledgerType: shared.TransactionTypeSupervisionBACSPayment.Key(),
			expectedPaymentDetails: shared.PaymentDetails{
				Amount:       55000,
				BankDate:     time.Date(2025, time.February, 10, 0, 0, 0, 0, time.UTC),
				CourtRef:     "54326543",
				LedgerType:   shared.TransactionTypeSupervisionBACSPayment.Key(),
				ReceivedDate: time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name:       "OPG BACS Record",
			record:     []string{"", "", "", "", "02/05/2025", "", "5.12", "", "", "", "12983476"},
			uploadType: shared.ReportTypeUploadPaymentsOPGBACS.Key(),
			bankDate:   shared.NewDate("2025-05-11"),
			ledgerType: shared.TransactionTypeOPGBACSPayment.Key(),
			expectedPaymentDetails: shared.PaymentDetails{
				Amount:       512,
				BankDate:     time.Date(2025, time.May, 11, 0, 0, 0, 0, time.UTC),
				CourtRef:     "12983476",
				LedgerType:   shared.TransactionTypeOPGBACSPayment.Key(),
				ReceivedDate: time.Date(2025, time.May, 2, 0, 0, 0, 0, time.UTC),
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name:       "SOP Unallocated Record",
			record:     []string{"12439823", "123.9"},
			uploadType: "SOP_UNALLOCATED",
			bankDate:   shared.NewDate("2025-06-11"),
			ledgerType: shared.TransactionTypeSOPUnallocatedPayment.Key(),
			expectedPaymentDetails: shared.PaymentDetails{
				Amount:       12390,
				BankDate:     time.Date(2025, time.June, 11, 0, 0, 0, 0, time.UTC),
				CourtRef:     "12439823",
				LedgerType:   shared.TransactionTypeSOPUnallocatedPayment.Key(),
				ReceivedDate: time.Date(2025, time.March, 31, 0, 0, 0, 0, time.UTC),
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name:                "Amount parse error",
			record:              []string{"87654321-abcdefg-garbleddata", "01-02-2025 11:33:55", "oops!"},
			uploadType:          shared.ReportTypeUploadPaymentsOnlineCard.Key(),
			bankDate:            shared.NewDate("2025-02-10"),
			ledgerType:          shared.TransactionTypeOnlineCardPayment.Key(),
			index:               1,
			expectedFailedLines: map[int]string{1: "AMOUNT_PARSE_ERROR"},
		},
		{
			name:                   "Date parse error",
			record:                 []string{"", "", "", "", "darn", "", "5.12", "", "", "", "12983476"},
			uploadType:             shared.ReportTypeUploadPaymentsOPGBACS.Key(),
			bankDate:               shared.NewDate("2025-05-11"),
			ledgerType:             shared.TransactionTypeOPGBACSPayment.Key(),
			index:                  0,
			expectedPaymentDetails: shared.PaymentDetails{},
			expectedFailedLines:    map[int]string{0: "DATE_PARSE_ERROR"},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			failedLines := make(map[int]string)

			paymentDetails := getPaymentDetails(tt.record, tt.uploadType, tt.bankDate, tt.ledgerType, tt.index, &failedLines)

			assert.Equal(t, tt.expectedPaymentDetails, paymentDetails)
			assert.Equal(t, tt.expectedFailedLines, failedLines)
		})
	}
}

func Test_parseAmount(t *testing.T) {
	tests := []struct {
		name         string
		stringAmount string
		wantAmount   int32
		wantErr      bool
	}{
		{
			name:         "No decimal",
			stringAmount: "500",
			wantAmount:   50000,
		},
		{
			name:         "Single decimal",
			stringAmount: "500.1",
			wantAmount:   50010,
		},
		{
			name:         "Two decimals",
			stringAmount: "500.12",
			wantAmount:   50012,
		},
		{
			name:         "Unable to parse",
			stringAmount: "hehe",
			wantErr:      true,
		},
		{
			name:         "Empty string",
			stringAmount: "",
			wantAmount:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, err := parseAmount(tt.stringAmount)
			assert.Equal(t, tt.wantAmount, amount)

			if tt.wantErr {
				assert.IsType(t, &strconv.NumError{}, err)
			} else {
				assert.Nil(t, err)
			}
		})
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
