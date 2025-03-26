package service

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type createdReversalAllocation struct {
	id               int
	ledgerAmount     int
	ledgerType       string
	ledgerStatus     string
	receivedDate     time.Time
	bankDate         time.Time
	allocationAmount int
	allocationStatus string
	invoiceId        int
}

func (suite *IntegrationSuite) Test_processReversals() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, 'failures', 'DEMANDED', NULL, '1111');",
		"INSERT INTO finance_client VALUES (2, 2, 'failures - replacement', 'DEMANDED', NULL, '2222');",
		"INSERT INTO invoice VALUES (1, 1, 1, 'AD', 'AD11223/19', '2023-04-01', '2025-03-31', 15000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (1, 'client-1-reverse-me', '2024-01-02 15:32:10', '', 1000, 'payment', 'MOTO CARD PAYMENT', 'CONFIRMED', 1, NULL, NULL, NULL, '2024-01-01', NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (1, 1, 1, '2024-01-02 15:32:10', 1000, 'ALLOCATED', NULL, '', '2024-01-01', NULL);",
		"ALTER SEQUENCE ledger_id_seq RESTART WITH 2;",
		"ALTER SEQUENCE ledger_allocation_id_seq RESTART WITH 2;",
	)

	dispatch := &mockDispatch{}
	s := NewService(seeder.Conn, dispatch, nil, nil, nil)

	tests := []struct {
		name                string
		records             [][]string
		expectedClientId    int
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
			expectedClientId: 1,
			////allocations: []createdReversalAllocation{
			////	{
			////		ledgerAmount:     -1000,
			////		ledgerType:       "MOTO CARD PAYMENT",
			////		ledgerStatus:     "CONFIRMED",
			////		receivedDate:     time.Date(2024, 01, 02, 15, 30, 27, 0, time.UTC),
			////		bankDate:         time.Date(2024, 01, 01, 0, 0, 0, 0, time.UTC),
			////		allocationAmount: 1000,
			////		allocationStatus: "ALLOCATED",
			////		invoiceId:        1,
			////	},
			////},
			expectedFailedLines: map[int]string{
				1: "NO_MATCHED_PAYMENT",
				2: "COURT_REF_MISMATCH",
				3: "NO_MATCHED_PAYMENT",
				4: "NO_MATCHED_PAYMENT",
				5: "NO_MATCHED_PAYMENT",
				6: "NO_MATCHED_PAYMENT",
			},
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
				`SELECT l.id, l.amount, l.type, l.status, l.datetime, l.bankdate, la.amount, la.status, la.invoice_id
						FROM ledger l
						JOIN ledger_allocation la ON l.id = la.ledger_id
					WHERE l.finance_client_id = $1 AND l.id > $2`, tt.expectedClientId, currentLedgerId)

			for rows.Next() {
				var r createdReversalAllocation
				_ = rows.Scan(&r.id, &r.ledgerAmount, &r.ledgerType, &r.ledgerStatus, &r.receivedDate, &r.bankDate, &r.allocationAmount, &r.allocationStatus, &r.invoiceId)
				allocations = append(allocations, r)
			}

			assert.Equal(t, tt.allocations, allocations)
		})
	}
}
