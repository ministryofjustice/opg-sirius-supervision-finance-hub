package service

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func (suite *IntegrationSuite) Test_processAdhocEvent() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (4, 4, 'invoice-4', 'DEMANDED', NULL, '1234567');",
		"INSERT INTO invoice VALUES (5, 4, 4, 'AD', 'AD11227/19', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (1, 'ref', '2024-01-01 15:30:27', '', 15000, 'payment', 'MOTO CARD PAYMENT', 'CONFIRMED', 4, NULL, NULL, NULL, '2024-01-01', NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (1, 1, 5, '2024-01-01 15:30:27', 15000, 'ALLOCATED', NULL, '', '2024-01-01', NULL);",
		"ALTER SEQUENCE ledger_id_seq RESTART WITH 2;",
		"ALTER SEQUENCE ledger_allocation_id_seq RESTART WITH 2;",
	)

	dispatch := &mockDispatch{}
	s := NewService(seeder.Conn, dispatch, nil, nil, nil)

	now := time.Now().UTC()
	todaysDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name                      string
		expectedPaymentDetails    shared.PaymentDetails
		expectedLedgerAllocations []createdLedgerAllocation
		expectedFinanceClientId   int
		errorExpected             error
	}{
		{
			name: "Has negative invoices",
			expectedPaymentDetails: shared.PaymentDetails{
				Amount: -500,
				ReceivedDate: pgtype.Timestamp{
					Time:             todaysDate,
					InfinityModifier: 0,
					Valid:            true,
				},
				CourtRef:   pgtype.Text{String: "AD11227/19", Valid: true},
				LedgerType: shared.TransactionTypeUnappliedPayment,
				BankDate: pgtype.Date{
					Time:             todaysDate,
					InfinityModifier: 0,
					Valid:            true,
				},
				CreatedBy: pgtype.Int4{
					Int32: 10,
					Valid: true,
				},
			},
			expectedLedgerAllocations: []createdLedgerAllocation{
				{
					-5000,
					"UNAPPLIED PAYMENT",
					"CONFIRMED",
					todaysDate,
					-5000,
					"ALLOCATED",
					5,
					0,
				},
			},
			expectedFinanceClientId: 4,
			errorExpected:           nil,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			var currentLedgerId int
			_ = seeder.QueryRow(suite.ctx, `SELECT MAX(id) FROM ledger`).Scan(&currentLedgerId)

			err := s.ProcessAdhocEvent(suite.ctx)
			assert.Equal(t, tt.errorExpected, err)

			var createdLedgerAllocations []createdLedgerAllocation

			rows, _ := seeder.Query(suite.ctx,
				`SELECT l.amount, l.type, l.status, l.datetime, la.amount, la.status, COALESCE(l.pis_number, 0), COALESCE(la.invoice_id, 0)
						FROM ledger l
						JOIN ledger_allocation la ON l.id = la.ledger_id
					WHERE l.finance_client_id = $1 AND l.id > $2`, tt.expectedFinanceClientId, currentLedgerId)

			for rows.Next() {
				var r createdLedgerAllocation
				_ = rows.Scan(&r.ledgerAmount, &r.ledgerType, &r.ledgerStatus, &r.datetime, &r.allocationAmount, &r.allocationStatus, &r.pisNumber, &r.invoiceId)
				createdLedgerAllocations = append(createdLedgerAllocations, r)
			}

			assert.Equal(t, tt.expectedLedgerAllocations, createdLedgerAllocations)
		})
	}
}
