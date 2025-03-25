package service

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
)

func (suite *IntegrationSuite) Test_processReversals() {
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
		expectedClientId          int
		expectedLedgerAllocations []createdLedgerAllocation
		expectedFailedLines       map[int]string
		want                      error
	}{
		//{
		//	name: "failure cases",
		//	records: [][]string{
		//		{"Payment type", "Current (errored) court reference", "New (correct) court reference", "Bank date", "Received date", "Amount", "PIS number (cheque only)"},
		//		{}, // incorrect payment type
		//		{}, // current court reference not found
		//		{}, // new court reference not found
		//		{}, // bank date does not match payment
		//		{}, // received date does not match payment
		//		{}, // amount does not match payment
		//		{}, // PIS number does not match payment (cheque only)
		//	},
		//	expectedClientId:          1,
		//	expectedLedgerAllocations: []createdLedgerAllocation{},
		//	expectedFailedLines:       map[int]string{},
		//},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			var failedLines map[int]string
			failedLines, err := s.ProcessPaymentReversals(suite.ctx, tt.records, shared.ReportTypeUploadPaymentsMOTOCard)
			assert.Equal(t, tt.want, err)
			assert.Equal(t, tt.expectedFailedLines, failedLines)

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
