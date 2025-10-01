package service

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_ProcessFailedDirectDebitCollections() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (2, 22, '', 'DEMANDED', NULL, 'reverse');",
		"INSERT INTO finance_client VALUES (3, 33, '', 'DEMANDED', NULL, 'reverse too');",

		// wrong date for payment but reverse on this invoice due to raised date
		"INSERT INTO invoice VALUES (2, 22, 2, 'AD', 'invoice-2', '2023-04-01', '2025-03-31', 15000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (2, 'wrong date/right client', '2025-09-02 15:32:10', '', 10000, 'payment 2', 'DIRECT DEBIT PAYMENT', 'CONFIRMED', 2, NULL, NULL, NULL, '2025-09-02', NULL, NULL, NULL, NULL, '2025-09-01', 1);",
		"INSERT INTO ledger_allocation VALUES (2, 2, 2, '2025-09-02 15:32:10', 10000, 'ALLOCATED', NULL, '', '2025-09-02', NULL);",

		// wrong type
		"INSERT INTO invoice VALUES (3, 22, 2, 'AD', 'invoice-3', '2023-04-01', '2025-03-31', 15000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (3, 'wrong type/right client', '2025-09-01 15:32:10', '', 10000, 'payment 3', 'MOTO CARD PAYMENT', 'CONFIRMED', 2, NULL, NULL, NULL, '2025-09-01', NULL, NULL, NULL, NULL, '2025-09-01', 1);",
		"INSERT INTO ledger_allocation VALUES (3, 3, 3, '2025-09-01 15:32:10', 10000, 'ALLOCATED', NULL, '', '2025-09-01', NULL);",

		// payment to reverse 1 but apply to invoice 2
		"INSERT INTO invoice VALUES (4, 22, 2, 'AD', 'invoice-4', '2023-04-01', '2025-03-31', 15000, NULL, '2024-03-31', NULL, '2024-02-01', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (4, 'match client 2', '2025-09-01 15:32:10', '', 10000, 'payment 4', 'DIRECT DEBIT PAYMENT', 'CONFIRMED', 2, NULL, NULL, NULL, '2025-09-01', NULL, NULL, NULL, NULL, '2025-09-01', 1);",
		"INSERT INTO ledger_allocation VALUES (4, 4, 4, '2025-09-01 15:32:10', 10000, 'ALLOCATED', NULL, '', '2025-09-01', NULL);",

		// payment to reverse 2
		"INSERT INTO invoice VALUES (5, 33, 3, 'AD', 'invoice-5', '2023-04-01', '2025-03-31', 20000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (5, 'match client 3', '2025-08-25 15:32:10', '', 10000, 'payment 5', 'DIRECT DEBIT PAYMENT', 'CONFIRMED', 3, NULL, NULL, NULL, '2025-08-25', NULL, NULL, NULL, NULL, '2025-09-01', 1);",
		"INSERT INTO ledger_allocation VALUES (5, 5, 5, '2025-08-25 15:32:10', 10000, 'ALLOCATED', NULL, '', '2025-08-25', NULL);",

		// already reversed
		"INSERT INTO ledger VALUES (6, 'duplicate/right client', '2025-08-28 15:32:10', '', 10000, 'payment 6', 'DIRECT DEBIT PAYMENT', 'CONFIRMED', 3, NULL, NULL, NULL, '2025-08-28', NULL, NULL, NULL, NULL, '2025-09-01', 1);",
		"INSERT INTO ledger_allocation VALUES (6, 6, 5, '2025-08-28 15:32:10', 10000, 'ALLOCATED', NULL, '', '2025-08-28', NULL);",
		"INSERT INTO ledger VALUES (7, 'existing reversal', '2025-08-31 15:32:10', '', -10000, 'payment 6 reversed', 'DIRECT DEBIT PAYMENT', 'CONFIRMED', 3, NULL, NULL, NULL, '2025-08-31', NULL, NULL, NULL, NULL, '2025-09-01', 1);",
		"INSERT INTO ledger_allocation VALUES (7, 7, 5, '2025-08-31 15:32:10', -10000, 'ALLOCATED', NULL, 'refer to payee', '2025-08-31', NULL);",

		"ALTER SEQUENCE ledger_id_seq RESTART WITH 8;",
		"ALTER SEQUENCE ledger_allocation_id_seq RESTART WITH 8;",
	)

	allpayMock := &mockAllpay{}
	collectionDate, _ := time.Parse("2006-01-02", "2025-09-01")
	govUKMock := &mockGovUK{workingDay: collectionDate.AddDate(0, 0, -10)}
	s := Service{store: store.New(seeder.Conn), allpay: allpayMock, govUK: govUKMock, tx: seeder.Conn}

	tests := []struct {
		name           string
		failedPayments allpay.FailedPayments
		apiError       error
		allocations    []createdReversalAllocation
		want           error
	}{
		{
			name:           "no failed payments",
			failedPayments: allpay.FailedPayments{},
		},
		{
			name: "failed payments processed with skips",
			failedPayments: allpay.FailedPayments{
				{ // wrong amount
					Amount:          99999,
					ClientReference: "reverse",
					CollectionDate:  "01/09/2025 11:22:33",
					ProcessedDate:   "10/09/2025 11:22:33",
					ReasonCode:      "xxx",
				},
				{ // wrong date
					Amount:          10000,
					ClientReference: "reverse",
					CollectionDate:  "31/08/2025 11:22:33",
					ProcessedDate:   "10/09/2025 11:22:33",
					ReasonCode:      "xxx",
				},
				{ // already processed
					Amount:          10000,
					ClientReference: "reverse too",
					CollectionDate:  "08/28/2025 11:22:33",
					ProcessedDate:   "31/08/2025 11:22:33",
					ReasonCode:      "xxx",
				},
				{
					Amount:          10000,
					ClientReference: "reverse",
					CollectionDate:  "01/09/2025 11:22:33",
					ProcessedDate:   "10/09/2025 11:22:33",
					ReasonCode:      "reason A",
				},
				{
					Amount:          10000,
					ClientReference: "reverse too",
					CollectionDate:  "25/08/2025 11:22:33",
					ProcessedDate:   "01/09/2025 11:22:33",
					ReasonCode:      "reason B",
				},
			},
			allocations: []createdReversalAllocation{
				{
					ledgerAmount:     -10000,
					ledgerType:       "DIRECT DEBIT PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 9, 10, 11, 22, 33, 0, time.UTC),
					bankDate:         time.Date(2025, 9, 10, 00, 00, 00, 0, time.UTC),
					allocationAmount: -10000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 2, Valid: true}, // payment will reverse the most recent invoice by raised date
					financeClientId:  2,
					notes:            "reason A",
				},
				{
					ledgerAmount:     -10000,
					ledgerType:       "DIRECT DEBIT PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 9, 01, 11, 22, 33, 0, time.UTC),
					bankDate:         time.Date(2025, 9, 01, 00, 00, 00, 0, time.UTC),
					allocationAmount: -10000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 5, Valid: true},
					financeClientId:  3,
					notes:            "reason B",
				},
			},
		},
		{
			name: "api error",
			failedPayments: allpay.FailedPayments{
				{
					Amount:          10000,
					ClientReference: "reverse",
					CollectionDate:  "2025-09-01",
					ProcessedDate:   "2025-09-10",
					ReasonCode:      "xxx",
				},
			},
			apiError: errors.New("api error"),
			want:     errors.New("api error"),
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			allpayMock.failedPayments = tt.failedPayments
			allpayMock.errs = map[string]error{"FetchFailedPayments": tt.apiError}
			var currentLedgerId int
			_ = seeder.QueryRow(suite.ctx, `SELECT MAX(id) FROM ledger`).Scan(&currentLedgerId)

			err := s.ProcessFailedDirectDebitCollections(suite.ctx, collectionDate)

			assert.Equal(t, tt.want, err)
			assert.Equal(t, allpay.FetchFailedPaymentsInput{
				To:   collectionDate,
				From: collectionDate.AddDate(0, 0, -10),
			}, allpayMock.lastCalledParams[0])

			var allocations []createdReversalAllocation

			rows, _ := seeder.Query(suite.ctx,
				`SELECT l.amount, l.type, l.status, l.datetime, l.bankdate, la.amount, la.status, l.finance_client_id, la.invoice_id, l.notes
						FROM ledger l
						LEFT JOIN ledger_allocation la ON l.id = la.ledger_id
					WHERE l.id > $1`, currentLedgerId)

			for rows.Next() {
				var r createdReversalAllocation
				_ = rows.Scan(&r.ledgerAmount, &r.ledgerType, &r.ledgerStatus, &r.receivedDate, &r.bankDate, &r.allocationAmount, &r.allocationStatus, &r.financeClientId, &r.invoiceId, &r.notes)
				allocations = append(allocations, r)
			}

			assert.Equal(t, tt.allocations, allocations)
		})
	}
}
