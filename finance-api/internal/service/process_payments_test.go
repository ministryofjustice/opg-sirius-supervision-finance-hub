package service

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"strconv"
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

func (suite *IntegrationSuite) Test_processPayments() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, 'invoice-1', 'DEMANDED', NULL, '1234');",
		"INSERT INTO finance_client VALUES (2, 2, 'invoice-2', 'DEMANDED', NULL, '12345');",
		"INSERT INTO finance_client VALUES (3, 3, 'invoice-3', 'DEMANDED', NULL, '123456');",
		"INSERT INTO finance_client VALUES (4, 4, 'invoice-4', 'DEMANDED', NULL, '1234567');",
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

	dispatch := &mockDispatch{}
	s := NewService(seeder.Conn, dispatch, nil, nil, nil)

	tests := []struct {
		name                      string
		records                   [][]string
		bankDate                  shared.Date
		expectedClientId          int
		expectedLedgerAllocations []createdLedgerAllocation
		expectedFailedLines       map[int]string
		want                      error
	}{
		{
			name: "Underpayment",
			records: [][]string{
				{"Ordercode", "Date", "Amount"},
				{
					"1234-1",
					"2024-01-01 10:15:39",
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
					time.Date(2024, 1, 1, 10, 15, 39, 0, time.UTC),
					10000,
					"ALLOCATED",
					1,
				},
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name: "Overpayment",
			records: [][]string{
				{"Ordercode", "Date", "Amount"},
				{
					"12345",
					"2024-01-01 15:30:27",
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
					time.Date(2024, 1, 1, 15, 30, 27, 0, time.UTC),
					10000,
					"ALLOCATED",
					2,
				},
				{
					25010,
					"MOTO CARD PAYMENT",
					"CONFIRMED",
					time.Date(2024, 1, 1, 15, 30, 27, 0, time.UTC),
					-15010,
					"UNAPPLIED",
					0,
				},
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name: "Underpayment with multiple invoices",
			records: [][]string{
				{"Ordercode", "Date", "Amount"},
				{
					"123456",
					"2024-01-01 15:30:27",
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
					time.Date(2024, 1, 1, 15, 30, 27, 0, time.UTC),
					5000,
					"ALLOCATED",
					3,
				},
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name: "failure cases",
			records: [][]string{
				{"Ordercode", "Date", "Amount"},
				{"1234567890", "2024-01-01 15:30:27", "50"}, // client not found
				{"1234567", "2024-01-01 15:30:27", "100"},   // duplicate
			},
			bankDate:         shared.NewDate("2024-01-01"),
			expectedClientId: 3,
			expectedFailedLines: map[int]string{
				1: "CLIENT_NOT_FOUND",
				2: "DUPLICATE_PAYMENT",
			},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			var currentLedgerId int
			_ = seeder.QueryRow(suite.ctx, `SELECT currval('ledger_id_seq')`).Scan(&currentLedgerId)

			var failedLines map[int]string
			failedLines, err := s.ProcessPayments(suite.ctx, tt.records, shared.ReportTypeUploadPaymentsMOTOCard, tt.bankDate)
			assert.Equal(t, tt.want, err)
			assert.Equal(t, tt.expectedFailedLines, failedLines)

			var createdLedgerAllocations []createdLedgerAllocation

			rows, _ := seeder.Query(suite.ctx,
				`SELECT l.amount, l.type, l.status, l.datetime, la.amount, la.status, la.invoice_id
						FROM ledger l
						JOIN ledger_allocation la ON l.id = la.ledger_id
					WHERE l.finance_client_id = $1 AND l.id > $2`, tt.expectedClientId, currentLedgerId)

			for rows.Next() {
				var r createdLedgerAllocation
				_ = rows.Scan(&r.ledgerAmount, &r.ledgerType, &r.ledgerStatus, &r.datetime, &r.allocationAmount, &r.allocationStatus, &r.invoiceId)
				createdLedgerAllocations = append(createdLedgerAllocations, r)
			}

			assert.Equal(t, tt.expectedLedgerAllocations, createdLedgerAllocations)
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
		name        string
		record      []string
		uploadType  shared.ReportUploadType
		bankDate    shared.Date
		ledgerType  string
		index       int
		expected    shared.PaymentDetails
		expectedErr map[int]string
	}{
		{
			name:       "MOTO Card Payment",
			record:     []string{"1234-1", "2024-01-01 10:15:39", "100.00"},
			uploadType: shared.ReportTypeUploadPaymentsMOTOCard,
			bankDate:   shared.NewDate("2024-01-17"),
			ledgerType: "MOTO_CARD_PAYMENT",
			index:      1,
			expected: shared.PaymentDetails{
				CourtRef:     "1234",
				ReceivedDate: time.Date(2024, 1, 1, 10, 15, 39, 0, time.UTC),
				Amount:       10000,
				BankDate:     time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC),
				LedgerType:   "MOTO_CARD_PAYMENT",
			},
			expectedErr: map[int]string{},
		},
		{
			name:       "Supervision BACS Payment",
			record:     []string{"", "", "", "", "01/01/2024", "", "200.00", "", "", "", "1234"},
			uploadType: shared.ReportTypeUploadPaymentsSupervisionBACS,
			bankDate:   shared.NewDate("2024-01-17"),
			ledgerType: "SUPERVISION_BACS_PAYMENT",
			index:      1,
			expected: shared.PaymentDetails{
				CourtRef:     "1234",
				ReceivedDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Amount:       20000,
				BankDate:     time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC),
				LedgerType:   "SUPERVISION_BACS_PAYMENT",
			},
			expectedErr: map[int]string{},
		},
		{
			name:       "Invalid Amount",
			record:     []string{"1234-1", "2024-01-01 10:15:39", "invalid"},
			uploadType: shared.ReportTypeUploadPaymentsMOTOCard,
			bankDate:   shared.NewDate("2024-01-17"),
			ledgerType: "MOTO_CARD_PAYMENT",
			index:      1,
			expected:   shared.PaymentDetails{},
			expectedErr: map[int]string{
				1: "AMOUNT_PARSE_ERROR",
			},
		},
		{
			name:       "Invalid Date",
			record:     []string{"1234-1", "invalid", "100.00"},
			uploadType: shared.ReportTypeUploadPaymentsMOTOCard,
			bankDate:   shared.NewDate("2024-01-17"),
			ledgerType: "MOTO_CARD_PAYMENT",
			index:      1,
			expected:   shared.PaymentDetails{},
			expectedErr: map[int]string{
				1: "DATE_TIME_PARSE_ERROR",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			failedLines := make(map[int]string)
			result := getPaymentDetails(tt.record, tt.uploadType, tt.bankDate, tt.ledgerType, tt.index, &failedLines)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedErr, failedLines)
		})
	}
}
