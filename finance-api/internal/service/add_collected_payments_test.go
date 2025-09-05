package service

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/stretchr/testify/assert"
)

type collectionLedger struct {
	paymentType  string
	amount       int
	ledgerID     int
	receivedDate time.Time
	bankDate     time.Time
}

func (suite *IntegrationSuite) Test_AddCollectedPayments() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 101, 'not-processed', 'DIRECT DEBIT', NULL, '1234');",
		"INSERT INTO finance_client VALUES (2, 201, 'processed-1', 'DIRECT DEBIT', NULL, '12345');",
		"INSERT INTO finance_client VALUES (3, 301, 'processed-2', 'DIRECT DEBIT', NULL, '12345');",
		"INSERT INTO invoice VALUES (1, 101, 1, 'AD', 'AD11223/19', '2023-04-01', '2025-03-31', 15000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO invoice VALUES (2, 201, 2, 'AD', 'AD11224/19', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO invoice VALUES (3, 301, 3, 'AD', 'AD11225/19', '2023-04-01', '2025-03-31', 15000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (1, 'ref', '2024-01-01 15:30:27', '', 10000, 'payment', 'MOTO CARD PAYMENT', 'CONFIRMED', 1, NULL, NULL, NULL, '2024-01-01', NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO pending_collection VALUES (1, 1, '2025-12-12', 15000, 'PENDING', NULL, '2025-08-21', 1);",
		"INSERT INTO pending_collection VALUES (2, 2, '2025-06-06', 10000, 'PENDING', NULL, '2025-08-21', 1);",
		"INSERT INTO pending_collection VALUES (3, 3, '2025-06-06', 15000, 'PENDING', NULL, '2025-08-21', 1);",
		"INSERT INTO pending_collection VALUES (4, 1, '2025-06-06', 10000, 'CANCELLED', NULL, '2025-08-21', 1);", // already processed (cancelled)
		"ALTER SEQUENCE ledger_id_seq RESTART WITH 2;",
	)

	tests := []struct {
		name                      string
		collectionDate            time.Time
		expectedLedgerAllocations []collectionLedger
		expectedErr               error
	}{
		{
			name:                      "No records for collection date",
			collectionDate:            time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedLedgerAllocations: []collectionLedger{},
		},
		{
			name:           "creates ledgers for collection date",
			collectionDate: time.Date(2025, 6, 6, 0, 0, 0, 0, time.UTC),
			expectedLedgerAllocations: []collectionLedger{
				{
					"DIRECT DEBIT PAYMENT",
					10000,
					2,
					time.Date(2025, 6, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 6, 6, 0, 0, 0, 0, time.UTC),
				},
				{
					"DIRECT DEBIT PAYMENT",
					15000,
					4,
					time.Date(2025, 6, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 6, 6, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			dispatch := &mockDispatch{}
			s := &Service{store: store.New(seeder.Conn), dispatch: dispatch, tx: seeder.Conn}

			err := s.AddCollectedPayments(suite.ctx, tt.collectionDate)
			assert.Equal(t, tt.expectedErr, err)

			ledgers := []collectionLedger{}
			rows, _ := seeder.Query(suite.ctx,
				`SELECT l.amount, l.type, l.id, l.datetime, l.bankdate
						FROM ledger l
						JOIN pending_collection pc ON pc.ledger_id = l.id
					WHERE pc.collection_date = $1`, tt.collectionDate)

			for rows.Next() {
				var r collectionLedger
				_ = rows.Scan(&r.amount, &r.paymentType, &r.ledgerID, &r.receivedDate, &r.bankDate)
				ledgers = append(ledgers, r)
			}

			assert.Equal(t, tt.expectedLedgerAllocations, ledgers)
		})
	}
}
