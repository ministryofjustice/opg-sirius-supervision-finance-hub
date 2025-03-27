package service

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type createdReversalAllocation struct {
	ledgerAmount     int
	ledgerType       string
	ledgerStatus     string
	receivedDate     time.Time
	bankDate         time.Time
	allocationAmount int
	allocationStatus string
	invoiceId        pgtype.Int4
	financeClientId  int
}

func (suite *IntegrationSuite) Test_processReversals() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		// test 1
		"INSERT INTO finance_client VALUES (1, 1, 'test 1', 'DEMANDED', NULL, '1111');",
		"INSERT INTO finance_client VALUES (2, 2, 'test 1 - replacement', 'DEMANDED', NULL, '2222');",
		"INSERT INTO invoice VALUES (1, 1, 1, 'AD', 'test 1 paid', '2023-04-01', '2025-03-31', 15000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (1, 'client-1-reverse-me', '2024-01-02 15:32:10', '', 1000, 'payment 1', 'MOTO CARD PAYMENT', 'CONFIRMED', 1, NULL, NULL, NULL, '2024-01-01', NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (1, 1, 1, '2024-01-02 15:32:10', 1000, 'ALLOCATED', NULL, '', '2024-01-01', NULL);",
		"INSERT INTO invoice VALUES (2, 2, 2, 'AD', 'test 1 replacement unpaid', '2023-04-01', '2025-03-31', 15000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",

		//test 2
		"INSERT INTO finance_client VALUES (3, 3, 'test 2', 'DEMANDED', NULL, '3333');",
		"INSERT INTO finance_client VALUES (4, 4, 'test 2 - replacement', 'DEMANDED', NULL, '4444');",
		"INSERT INTO invoice VALUES (3, 3, 3, 'AD', 'test 2 paid', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO invoice VALUES (4, 3, 3, 'AD', 'test 2 partially paid with payment', '2023-05-01', '2025-04-01', 10000, NULL, '2024-04-01', NULL, '2024-04-01', NULL, NULL, NULL, '2024-04-01 00:00:00', '99');",
		"INSERT INTO invoice VALUES (5, 3, 3, 'AD', 'test 2 unpaid with payment', '2023-06-01', '2025-05-01', 10000, NULL, '2025-05-01', NULL, '2025-05-01', NULL, NULL, NULL, '2025-05-01 00:00:00', '99');",
		"INSERT INTO ledger VALUES (2, 'test 2', '2025-01-02 15:32:10', '', 15000, 'payment 2', 'ONLINE CARD PAYMENT', 'CONFIRMED', 3, NULL, NULL, NULL, '2025-01-02', NULL, NULL, NULL, NULL, '2025-01-02', 1);",
		"INSERT INTO ledger_allocation VALUES (2, 2, 3, '2025-01-02 15:32:10', 10000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",
		"INSERT INTO ledger_allocation VALUES (3, 2, 4, '2025-01-02 15:32:10', 5000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",
		// second payment received after the payment being reversed
		"INSERT INTO ledger VALUES (3, 'test 2 - second payment', '2025-01-03 15:32:10', '', 10000, 'payment 2 - not reversed', 'MOTO CARD PAYMENT', 'CONFIRMED', 3, NULL, NULL, NULL, '2025-01-03', NULL, NULL, NULL, NULL, '2025-01-03', 1);",
		"INSERT INTO ledger_allocation VALUES (4, 3, 4, '2025-01-03 15:32:10', 5000, 'ALLOCATED', NULL, '', '2025-01-03', NULL);",
		"INSERT INTO ledger_allocation VALUES (5, 3, 5, '2025-01-03 15:32:10', 5000, 'ALLOCATED', NULL, '', '2025-01-03', NULL);",

		// test 3
		"INSERT INTO finance_client VALUES (5, 5, 'test 3', 'DEMANDED', NULL, '5555');",
		"INSERT INTO finance_client VALUES (6, 6, 'test 3 - replacement', 'DEMANDED', NULL, '6666');",
		"INSERT INTO invoice VALUES (6, 5, 5, 'AD', 'test 3 paid', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (4, 'test 3', '2025-01-02 15:32:10', '', 15000, 'payment 3', 'ONLINE CARD PAYMENT', 'CONFIRMED', 5, NULL, NULL, NULL, '2025-01-02', NULL, NULL, NULL, NULL, '2025-01-02', 1);",
		"INSERT INTO ledger_allocation VALUES (6, 4, 6, '2025-01-02 15:32:10', 10000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",
		"INSERT INTO ledger_allocation VALUES (7, 4, NULL, '2025-01-02 15:32:10', -5000, 'UNAPPLIED', NULL, '', '2025-01-02', NULL);",
		"INSERT INTO invoice VALUES (7, 6, 6, 'AD', 'test 3 replacement', '2023-04-01', '2025-03-31', 15000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",

		"ALTER SEQUENCE ledger_id_seq RESTART WITH 5;",
		"ALTER SEQUENCE ledger_allocation_id_seq RESTART WITH 8;",
	)

	dispatch := &mockDispatch{}
	s := NewService(seeder.Conn, dispatch, nil, nil, nil)

	tests := []struct {
		name                string
		records             [][]string
		allocations         []createdReversalAllocation
		expectedFailedLines map[int]string
		want                error
	}{
		{
			name: "failure cases with eventual success",
			records: [][]string{
				{"Payment type", "Current (errored) court reference", "New (correct) court reference", "Bank date", "Received date", "Amount", "PIS number (cheque only)"},
				{"MOTO CARD PAYMENT", "0000", "2222", "2024-01-01", "2024-01-02", "10.00", ""},   // current court reference not found
				{"MOTO CARD PAYMENT", "1111", "0000", "2024-01-01", "2024-01-02", "10.00", ""},   // new court reference not found
				{"ONLINE CARD PAYMENT", "1111", "2222", "2024-01-01", "2024-01-02", "10.00", ""}, // incorrect payment type
				{"MOTO CARD PAYMENT", "1111", "2222", "2024-01-12", "2024-01-02", "10.00", ""},   // bank date does not match payment
				{"MOTO CARD PAYMENT", "1111", "2222", "2024-01-01", "2024-01-12", "10.00", ""},   // received date does not match payment
				{"MOTO CARD PAYMENT", "1111", "2222", "2024-01-01", "2024-01-02", "10.01", ""},   // amount does not match payment
				{"MOTO CARD PAYMENT", "1111", "2222", "2024-01-01", "2024-01-02", "10.00", ""},   // successful match
			},
			allocations: []createdReversalAllocation{
				{
					ledgerAmount:     -1000,
					ledgerType:       "MOTO CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2024, 01, 02, 00, 00, 00, 0, time.UTC),
					bankDate:         time.Date(2024, 01, 01, 0, 0, 0, 0, time.UTC),
					allocationAmount: -1000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 1, Valid: true},
					financeClientId:  1,
				},
				{
					ledgerAmount:     1000,
					ledgerType:       "MOTO CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2024, 01, 02, 00, 00, 00, 0, time.UTC),
					bankDate:         time.Date(2024, 01, 01, 0, 0, 0, 0, time.UTC),
					allocationAmount: 1000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 2, Valid: true},
					financeClientId:  2,
				},
			},
			expectedFailedLines: map[int]string{
				1: "NO_MATCHED_PAYMENT",
				2: "COURT_REF_MISMATCH",
				3: "NO_MATCHED_PAYMENT",
				4: "NO_MATCHED_PAYMENT",
				5: "NO_MATCHED_PAYMENT",
				6: "NO_MATCHED_PAYMENT",
			},
		},
		{
			name: "original payment over two invoices applied to client with overpayment",
			records: [][]string{
				{"Payment type", "Current (errored) court reference", "New (correct) court reference", "Bank date", "Received date", "Amount", "PIS number (cheque only)"},
				{"ONLINE CARD PAYMENT", "3333", "4444", "2025-01-02", "2025-01-02", "150.00", ""},
			},
			allocations: []createdReversalAllocation{
				{
					ledgerAmount:     -15000,
					ledgerType:       "ONLINE CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: -5000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 5, Valid: true},
					financeClientId:  3,
				},
				{
					ledgerAmount:     -15000,
					ledgerType:       "ONLINE CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: -10000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 4, Valid: true},
					financeClientId:  3,
				},
				{
					ledgerAmount:     15000,
					ledgerType:       "ONLINE CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: -15000, // unapply so a negative amount is allocated
					allocationStatus: "UNAPPLIED",
					invoiceId:        pgtype.Int4{}, // no invoice on replacement client so unapplied as overpayment
					financeClientId:  4,
				},
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name: "errored client in credit",
			records: [][]string{
				{"Payment type", "Current (errored) court reference", "New (correct) court reference", "Bank date", "Received date", "Amount", "PIS number (cheque only)"},
				{"ONLINE CARD PAYMENT", "5555", "6666", "2025-01-02", "2025-01-02", "150.00", ""},
			},
			allocations: []createdReversalAllocation{
				{
					ledgerAmount:     -15000,
					ledgerType:       "ONLINE CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: -10000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 6, Valid: true},
					financeClientId:  5,
				},
				{
					ledgerAmount:     -15000,
					ledgerType:       "ONLINE CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: 5000, // positive unapply to reverse the existing credit balance
					allocationStatus: "UNAPPLIED",
					invoiceId:        pgtype.Int4{},
					financeClientId:  5,
				},
				{
					ledgerAmount:     15000,
					ledgerType:       "ONLINE CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: 15000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 7, Valid: true},
					financeClientId:  6,
				},
			},
			expectedFailedLines: map[int]string{},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			var currentLedgerId int
			_ = seeder.QueryRow(suite.ctx, `SELECT MAX(id) FROM ledger`).Scan(&currentLedgerId)

			var failedLines map[int]string
			failedLines, err := s.ProcessPaymentReversals(suite.ctx, tt.records, shared.ReportTypeUploadMisappliedPayments)
			assert.Equal(t, tt.want, err)
			assert.Equal(t, tt.expectedFailedLines, failedLines)

			var allocations []createdReversalAllocation

			rows, _ := seeder.Query(suite.ctx,
				`SELECT l.amount, l.type, l.status, l.datetime, l.bankdate, la.amount, la.status, l.finance_client_id, la.invoice_id
						FROM ledger l
						LEFT JOIN ledger_allocation la ON l.id = la.ledger_id
					WHERE l.id > $1`, currentLedgerId)

			for rows.Next() {
				var r createdReversalAllocation
				_ = rows.Scan(&r.ledgerAmount, &r.ledgerType, &r.ledgerStatus, &r.receivedDate, &r.bankDate, &r.allocationAmount, &r.allocationStatus, &r.financeClientId, &r.invoiceId)
				allocations = append(allocations, r)
			}

			assert.Equal(t, tt.allocations, allocations)
		})
	}
}
