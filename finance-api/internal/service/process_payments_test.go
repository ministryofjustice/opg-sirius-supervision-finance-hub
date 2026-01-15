package service

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

type createdLedgerAllocation struct {
	ledgerAmount     int
	ledgerType       string
	ledgerStatus     string
	datetime         time.Time
	allocationAmount int
	allocationStatus string
	invoiceId        int
	pisNumber        int
}

func (suite *IntegrationSuite) Test_processPayments() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, 'invoice-1', 'DEMANDED', NULL, '1234');",
		"INSERT INTO finance_client VALUES (2, 2, 'invoice-2', 'DEMANDED', NULL, '12345');",
		"INSERT INTO finance_client VALUES (3, 3, 'invoice-3', 'DEMANDED', NULL, '123456');",
		"INSERT INTO finance_client VALUES (4, 4, 'invoice-4', 'DEMANDED', NULL, '1234567');",
		"INSERT INTO finance_client VALUES (5, 5, 'duplicate-1', 'DEMANDED', NULL, '12345678');",
		"INSERT INTO finance_client VALUES (6, 6, 'duplicate-2', 'DEMANDED', NULL, '12345678');",
		"INSERT INTO invoice VALUES (1, 1, 1, 'AD', 'AD11223/19', '2023-04-01', '2025-03-31', 15000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO invoice VALUES (2, 2, 2, 'AD', 'AD11224/19', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO invoice VALUES (3, 3, 3, 'AD', 'AD11225/19', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO invoice VALUES (4, 3, 3, 'AD', 'AD11226/19', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO invoice VALUES (5, 4, 4, 'AD', 'AD11227/19', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (1, 'ref', '2024-01-01 15:30:27', '', 10000, 'payment', 'MOTO CARD PAYMENT', 'CONFIRMED', 4, NULL, NULL, NULL, '2024-01-01', NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (1, 1, 5, '2024-01-01 15:30:27', 10000, 'ALLOCATED', NULL, '', '2024-01-01', NULL);",
		"ALTER SEQUENCE ledger_id_seq RESTART WITH 2;",
		"ALTER SEQUENCE ledger_allocation_id_seq RESTART WITH 2;",
	)

	tests := []struct {
		name                      string
		records                   [][]string
		paymentType               shared.ReportUploadType
		bankDate                  shared.Date
		pisNumber                 int
		expectedClientId          int
		expectedLedgerAllocations []createdLedgerAllocation
		expectedFailedLines       map[int]string
		expectedDispatch          any
		ledgerCount               int
		want                      error
	}{
		{
			name: "Underpayment",
			records: [][]string{
				{"9800000000000000000", "1234", "100", "D", "01/01/2024"},
			},
			paymentType:      shared.ReportTypeUploadDirectDebitsCollections,
			bankDate:         shared.NewDate("2024-01-17"),
			expectedClientId: 1,
			ledgerCount:      1,
			expectedLedgerAllocations: []createdLedgerAllocation{
				{
					10000,
					"DIRECT DEBIT PAYMENT",
					"CONFIRMED",
					time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					10000,
					"ALLOCATED",
					1,
					-1,
				},
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name: "Overpayment",
			records: [][]string{
				{"Case number (confirmed on Sirius)", "Cheque number", "Cheque Value (£)", "Comments", "Date in Bank"},
				{"12345", "54321", "250.10", "", "01/01/2024"},
			},
			paymentType:      shared.ReportTypeUploadPaymentsSupervisionCheque,
			bankDate:         shared.NewDate("2024-01-17"),
			pisNumber:        150,
			expectedClientId: 2,
			ledgerCount:      1,
			expectedLedgerAllocations: []createdLedgerAllocation{
				{
					25010,
					"SUPERVISION CHEQUE PAYMENT",
					"CONFIRMED",
					time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					10000,
					"ALLOCATED",
					2,
					150,
				},
				{
					25010,
					"SUPERVISION CHEQUE PAYMENT",
					"CONFIRMED",
					time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					-15010,
					"UNAPPLIED",
					-1,
					150,
				},
			},
			expectedFailedLines: map[int]string{},
			expectedDispatch:    event.CreditOnAccount{ClientID: 2, CreditRemaining: 15010},
		},
		{
			name: "Underpayment with multiple invoices",
			records: [][]string{
				{"Ordercode", "Date", "Amount"},
				{"123456", "01/01/2024", "50"},
			},
			paymentType:      shared.ReportTypeUploadPaymentsMOTOCard,
			bankDate:         shared.NewDate("2024-01-17"),
			expectedClientId: 3,
			ledgerCount:      1,
			expectedLedgerAllocations: []createdLedgerAllocation{
				{
					5000,
					"MOTO CARD PAYMENT",
					"CONFIRMED",
					time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					5000,
					"ALLOCATED",
					3,
					-1,
				},
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name: "failure cases",
			records: [][]string{
				{"Ordercode", "Date", "Amount"},
				{"1234567890", "01/01/2024", "50"}, // client not found
				{"1234567", "01/01/2024", "100"},   // duplicate
			},
			paymentType:      shared.ReportTypeUploadPaymentsMOTOCard,
			bankDate:         shared.NewDate("2024-01-01"),
			expectedClientId: 3,
			expectedFailedLines: map[int]string{
				1: "CLIENT_NOT_FOUND",
				2: "DUPLICATE_PAYMENT",
			},
		},
		{
			name: "duplicate ledger prevention",
			records: [][]string{
				{"Ordercode", "Date", "Amount"},
				{"12345678", "01/01/2024", "50"},
			},
			paymentType:         shared.ReportTypeUploadPaymentsMOTOCard,
			bankDate:            shared.NewDate("2024-01-01"),
			expectedClientId:    3,
			expectedFailedLines: map[int]string{},
			ledgerCount:         1,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			dispatch := &mockDispatch{}
			s := Service{store: store.New(seeder.Conn), dispatch: dispatch, tx: seeder.Conn}

			var currentLedgerId int
			_ = seeder.QueryRow(suite.ctx, `SELECT MAX(id) FROM ledger`).Scan(&currentLedgerId)

			var failedLines map[int]string
			failedLines, err := s.ProcessPayments(suite.ctx, tt.records, tt.paymentType, tt.bankDate, tt.pisNumber)
			assert.Equal(t, tt.want, err)
			assert.Equal(t, tt.expectedFailedLines, failedLines)

			var createdLedgerAllocations []createdLedgerAllocation

			rows, _ := seeder.Query(suite.ctx,
				`SELECT l.amount, l.type, l.status, l.datetime, COALESCE(la.amount, -1), COALESCE(la.status, 'NOT_SET'), COALESCE(l.pis_number, -1), COALESCE(la.invoice_id, -1)
						FROM ledger l
						LEFT JOIN ledger_allocation la ON l.id = la.ledger_id
					WHERE l.finance_client_id = $1 AND l.id > $2`, tt.expectedClientId, currentLedgerId)

			for rows.Next() {
				var r createdLedgerAllocation
				_ = rows.Scan(&r.ledgerAmount, &r.ledgerType, &r.ledgerStatus, &r.datetime, &r.allocationAmount, &r.allocationStatus, &r.pisNumber, &r.invoiceId)
				createdLedgerAllocations = append(createdLedgerAllocations, r)
			}

			assert.Equal(t, tt.expectedLedgerAllocations, createdLedgerAllocations)

			var ledgerCount int
			_ = seeder.QueryRow(suite.ctx,
				`SELECT COUNT(*) FROM ledger WHERE id > $1`, currentLedgerId).Scan(&ledgerCount)
			assert.Equal(t, tt.ledgerCount, ledgerCount)

			assert.Equal(t, tt.expectedDispatch, dispatch.event)
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

func Test_getPaymentDetails(t *testing.T) {
	tests := []struct {
		name                   string
		record                 []string
		uploadType             shared.ReportUploadType
		pisNumber              int
		index                  int
		failedLines            map[int]string
		expectedPaymentDetails shared.PaymentDetails
		expectedFailedLines    map[int]string
	}{
		{
			name:       "Moto card",
			record:     []string{"12345678", "02/01/2025", "320.00"},
			uploadType: shared.ReportTypeUploadPaymentsMOTOCard,
			index:      0,
			expectedPaymentDetails: shared.PaymentDetails{
				Amount: 32000,
				ReceivedDate: pgtype.Timestamp{
					Time:             time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
					InfinityModifier: 0,
					Valid:            true,
				},
				CourtRef:   pgtype.Text{String: "12345678", Valid: true},
				LedgerType: shared.TransactionTypeMotoCardPayment,
				BankDate: pgtype.Date{
					Time:             time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					InfinityModifier: 0,
					Valid:            true,
				},
				CreatedBy: pgtype.Int4{
					Int32: 10,
					Valid: true,
				},
			},
		},
		{
			name:       "BACS",
			record:     []string{"", "", "", "", "01/06/2024", "", "20.50", "", "", "", "87654321"},
			uploadType: shared.ReportTypeUploadPaymentsSupervisionBACS,
			index:      0,
			expectedPaymentDetails: shared.PaymentDetails{
				Amount: 2050,
				ReceivedDate: pgtype.Timestamp{
					Time:             time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
					InfinityModifier: 0,
					Valid:            true,
				},
				CourtRef:   pgtype.Text{String: "87654321", Valid: true},
				LedgerType: shared.TransactionTypeSupervisionBACSPayment,
				BankDate: pgtype.Date{
					Time:             time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					InfinityModifier: 0,
					Valid:            true,
				},
				CreatedBy: pgtype.Int4{
					Int32: 10,
					Valid: true,
				},
			},
		},
		{
			name:       "CHEQUE",
			record:     []string{"23145746", "", "541.02", "", "31/10/2024"},
			uploadType: shared.ReportTypeUploadPaymentsSupervisionCheque,
			pisNumber:  123,
			index:      0,
			expectedPaymentDetails: shared.PaymentDetails{
				Amount: 54102,
				ReceivedDate: pgtype.Timestamp{
					Time:             time.Date(2024, 10, 31, 0, 0, 0, 0, time.UTC),
					InfinityModifier: 0,
					Valid:            true,
				},
				CourtRef:   pgtype.Text{String: "23145746", Valid: true},
				LedgerType: shared.TransactionTypeSupervisionChequePayment,
				BankDate: pgtype.Date{
					Time:             time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					InfinityModifier: 0,
					Valid:            true,
				},
				PisNumber: pgtype.Int4{
					Int32: 123,
					Valid: true,
				},
				CreatedBy: pgtype.Int4{
					Int32: 10,
					Valid: true,
				},
			},
		},
		{
			name:       "DIRECT DEBIT",
			record:     []string{"9800000000000000000", "012345678       ", "   200.92", "D", "05/03/2025"},
			uploadType: shared.ReportTypeUploadDirectDebitsCollections,
			index:      0,
			expectedPaymentDetails: shared.PaymentDetails{
				Amount: 20092,
				ReceivedDate: pgtype.Timestamp{
					Time:             time.Date(2025, 03, 05, 0, 0, 0, 0, time.UTC),
					InfinityModifier: 0,
					Valid:            true,
				},
				CourtRef:   pgtype.Text{String: "012345678", Valid: true},
				LedgerType: shared.TransactionTypeDirectDebitPayment,
				BankDate: pgtype.Date{
					Time:             time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					InfinityModifier: 0,
					Valid:            true,
				},
				CreatedBy: pgtype.Int4{
					Int32: 10,
					Valid: true,
				},
			},
		},
		{
			name:       "SOP unallocated",
			record:     []string{"012345678", "200.92"},
			uploadType: shared.ReportTypeUploadSOPUnallocated,
			index:      0,
			expectedPaymentDetails: shared.PaymentDetails{
				Amount: 20092,
				ReceivedDate: pgtype.Timestamp{
					Time:             time.Date(2025, 03, 31, 0, 0, 0, 0, time.UTC),
					InfinityModifier: 0,
					Valid:            true,
				},
				CourtRef:   pgtype.Text{String: "012345678", Valid: true},
				LedgerType: shared.TransactionTypeSOPUnallocatedPayment,
				BankDate: pgtype.Date{
					Time:             time.Date(2025, 03, 31, 0, 0, 0, 0, time.UTC),
					InfinityModifier: 0,
					Valid:            true,
				},
				CreatedBy: pgtype.Int4{
					Int32: 10,
					Valid: true,
				},
			},
		},
		{
			name:                "Amount parse error returns failed line",
			record:              []string{"23145746", "2024-01-01 00:00:00", "five hundred pounds!!!"},
			uploadType:          shared.ReportTypeUploadPaymentsMOTOCard,
			index:               0,
			failedLines:         map[int]string{},
			expectedFailedLines: map[int]string{0: "AMOUNT_PARSE_ERROR"},
		},
		{
			name:                "Date parse error returns failed line",
			record:              []string{"23145746", "", "200", "", "yesterday"},
			uploadType:          shared.ReportTypeUploadPaymentsSupervisionCheque,
			index:               0,
			failedLines:         map[int]string{},
			expectedFailedLines: map[int]string{0: "DATE_PARSE_ERROR"},
		},
		{
			name:                "Failed line adds to existing failed lines",
			record:              []string{"23145746", "yesterday", "200"},
			uploadType:          shared.ReportTypeUploadPaymentsMOTOCard,
			index:               1,
			failedLines:         map[int]string{0: "AMOUNT_PARSE_ERROR"},
			expectedFailedLines: map[int]string{0: "AMOUNT_PARSE_ERROR", 1: "DATE_PARSE_ERROR"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := auth.Context{
				Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test")),
				User:    &shared.User{ID: 10},
			}
			paymentDetails := getPaymentDetails(ctx, tt.record, tt.uploadType, shared.NewDate("01/01/2025"), tt.pisNumber, tt.index, &tt.failedLines)
			assert.Equal(t, tt.expectedFailedLines, tt.failedLines)
			assert.Equal(t, tt.expectedPaymentDetails, paymentDetails)
		})
	}
}

func Test_safeRead(t *testing.T) {
	type args struct {
		record []string
		index  int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "safe",
			args: args{
				record: []string{"abc"},
				index:  0,
			},
			want: "abc",
		},
		{
			name: "unsafe",
			args: args{
				record: []string{"abc"},
				index:  1000,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, safeRead(tt.args.record, tt.args.index), "safeRead(%v, %v)", tt.args.record, tt.args.index)
		})
	}
}
