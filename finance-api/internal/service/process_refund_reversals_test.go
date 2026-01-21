package service

import (
	"testing"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_processRefundReversals() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	bankDate := shared.NewDate("2024-12-12")

	seeder.SeedData(
        // Seed persons table to match finance_client court refs due to join in updated CreateLedgerForCourtRef query
	    "INSERT INTO public.persons VALUES (11, NULL, NULL, NULL, '12345678', NULL, NULL, NULL, false, false, NULL, NULL, 'Client', 'ACTIVE');",
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DEMANDED', NULL, '12345678');",
		"INSERT INTO ledger VALUES (1, 'payment-1', '2024-01-01 15:30:27', '', 5000, 'payment', 'SUPERVISION DIRECT DEBIT', 'CONFIRMED', 1, NULL, NULL, NULL, '2024-01-01', NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (1, 1, NULL, '2024-01-01 15:30:27', -5000, 'UNAPPLIED', NULL, '', '2024-01-01', NULL);",
		"INSERT INTO refund VALUES (1, 1, '2024-01-06', 5000, 'APPROVED', 'A fulfilled refund', 99, '2024-06-01 00:00:00', 99, '2024-06-02 00:00:00', '2024-06-03 00:00:00', NULL, '2024-06-05 00:00:00')",

		"INSERT INTO ledger VALUES (2, 'refund-1', '2024-06-05 15:30:27', '', -5000, 'refund', 'REFUND', 'CONFIRMED', 1, NULL, NULL, NULL, '2024-06-05', NULL, NULL, NULL, NULL, '2024-06-05', 1);",
		"INSERT INTO ledger_allocation VALUES (2, 2, NULL, '2024-06-05 15:30:27', 5000, 'REAPPLIED', NULL, '', '2024-06-05', NULL);",
		"INSERT INTO invoice VALUES (1, 11, 1, 's3', 'new invoice', '2023-04-01', '2025-03-31', 1000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",

		// already processed reversal
		"INSERT INTO public.persons VALUES (22, NULL, NULL, NULL, '87654321', NULL, NULL, NULL, false, false, NULL, NULL, 'Client', 'ACTIVE');",
		"INSERT INTO finance_client VALUES (2, 22, '8765', 'DEMANDED', NULL, '87654321');",
		"INSERT INTO ledger VALUES (3, 'payment-2', '2024-01-01 15:30:27', '', 5000, 'payment', 'SUPERVISION DIRECT DEBIT', 'CONFIRMED', 2, NULL, NULL, NULL, '2024-01-01', NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (3, 3, NULL, '2024-01-01 15:30:27', -5000, 'UNAPPLIED', NULL, '', '2024-01-01', NULL);",
		"INSERT INTO refund VALUES (2, 2, '2024-01-06', 5000, 'APPROVED', 'A fulfilled refund', 99, '2024-06-01 00:00:00', 99, '2024-06-02 00:00:00', '2024-06-03 00:00:00', NULL, '2024-06-05 00:00:00')",

		"INSERT INTO ledger VALUES (4, 'refund-2', '2024-06-05 15:30:27', '', -5000, 'refund', 'REFUND', 'CONFIRMED', 2, NULL, NULL, NULL, '2024-06-05', NULL, NULL, NULL, NULL, '2024-06-05', 1);",
		"INSERT INTO ledger_allocation VALUES (4, 4, NULL, '2024-06-05 15:30:27', 5000, 'REAPPLIED', NULL, '', '2024-06-05', NULL);",

		"INSERT INTO ledger VALUES (5, 'reversal-1', '2024-12-12 15:30:27', '', 5000, 'refund reversal', 'REFUND', 'CONFIRMED', 2, NULL, NULL, NULL, '2024-12-12', NULL, NULL, NULL, NULL, '2024-12-12', 1);",
		"INSERT INTO ledger_allocation VALUES (5, 5, NULL, '2024-12-12 15:30:27', -5000, 'UNAPPLIED', NULL, '', '2024-12-12', NULL);",

		"ALTER SEQUENCE ledger_id_seq RESTART WITH 6;",
		"ALTER SEQUENCE ledger_allocation_id_seq RESTART WITH 6;",
	)

	tests := []struct {
		name                      string
		records                   [][]string
		bankDate                  shared.Date
		expectedClientId          int
		expectedLedgerAllocations []createdLedgerAllocation
		expectedFailedLines       map[int]string
		expectedDispatch          any
		want                      error
	}{
		{
			name: "refund reversal",
			records: [][]string{
				{"Court reference", "Amount", "Bank date"},
				{"12345678", "50.00", "05/06/2024"},
			},
			bankDate:         bankDate,
			expectedClientId: 1,
			expectedLedgerAllocations: []createdLedgerAllocation{
				{
					ledgerAmount:     5000,
					ledgerType:       "REFUND",
					ledgerStatus:     "CONFIRMED",
					datetime:         bankDate.Time,
					allocationAmount: 1000,
					allocationStatus: "ALLOCATED",
					invoiceId:        1,
				},
				{
					ledgerAmount:     5000,
					ledgerType:       "REFUND",
					ledgerStatus:     "CONFIRMED",
					datetime:         bankDate.Time,
					allocationAmount: -4000,
					allocationStatus: "UNAPPLIED",
					invoiceId:        -1,
				},
			},
			expectedFailedLines: map[int]string{},
			expectedDispatch:    event.CreditOnAccount{ClientID: 11, CreditRemaining: 4000},
		},
		{
			name: "can't find refund to reverse - court reference",
			records: [][]string{
				{"Court reference", "Amount", "Bank date"},
				{"99999999", "50.00", "05/06/2024"},
			},
			bankDate:            bankDate,
			expectedClientId:    99,
			expectedFailedLines: map[int]string{1: "REFUND_NOT_FOUND_FOR_REVERSAL"},
		},
		{
			name: "can't find refund to reverse - amount",
			records: [][]string{
				{"Court reference", "Amount", "Bank date"},
				{"12345678", "50.05", "05/06/2024"},
			},
			bankDate:            bankDate,
			expectedClientId:    1,
			expectedFailedLines: map[int]string{1: "REFUND_NOT_FOUND_FOR_REVERSAL"},
		},
		{
			name: "can't find refund to reverse - date",
			records: [][]string{
				{"Court reference", "Amount", "Bank date"},
				{"12345678", "50.00", "06/07/2024"},
			},
			bankDate:            bankDate,
			expectedClientId:    1,
			expectedFailedLines: map[int]string{1: "REFUND_NOT_FOUND_FOR_REVERSAL"},
		},
		{
			name: "refund reversal already processed",
			records: [][]string{
				{"Court reference", "Amount", "Bank date"},
				{"87654321", "50.00", "05/06/2024"},
			},
			bankDate:            bankDate,
			expectedClientId:    2,
			expectedFailedLines: map[int]string{1: "DUPLICATE_PAYMENT"},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			dispatch := &mockDispatch{}
			s := Service{store: store.New(seeder.Conn), dispatch: dispatch, tx: seeder.Conn}

			var currentLedgerId int
			_ = seeder.QueryRow(suite.ctx, `SELECT MAX(id) FROM ledger`).Scan(&currentLedgerId)

			var failedLines map[int]string
			failedLines, err := s.ProcessRefundReversals(suite.ctx, tt.records, tt.bankDate)
			assert.Equal(t, tt.want, err)
			assert.Equal(t, tt.expectedFailedLines, failedLines)

			var createdLedgerAllocations []createdLedgerAllocation

			rows, err := seeder.Query(suite.ctx,
				`SELECT l.amount, l.type, l.status, l.datetime, COALESCE(la.amount, -1), COALESCE(la.status, 'NOT_SET'), COALESCE(la.invoice_id, -1)
						FROM ledger l
						LEFT JOIN ledger_allocation la ON l.id = la.ledger_id
					WHERE l.finance_client_id = $1 AND l.id > $2`, tt.expectedClientId, currentLedgerId)

			assert.NoError(t, err)

			for rows.Next() {
				var r createdLedgerAllocation
				err = rows.Scan(&r.ledgerAmount, &r.ledgerType, &r.ledgerStatus, &r.datetime, &r.allocationAmount, &r.allocationStatus, &r.invoiceId)
				assert.NoError(t, err)
				createdLedgerAllocations = append(createdLedgerAllocations, r)
			}

			assert.Equal(t, tt.expectedLedgerAllocations, createdLedgerAllocations)
			assert.Equal(t, tt.expectedDispatch, dispatch.event)
		})
	}
}
