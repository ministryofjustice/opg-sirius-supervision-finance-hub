package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func (suite *IntegrationSuite) Test_processFulfilledRefunds() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (99, 1, 'ian-test', 'DEMANDED', NULL, '12345678');",
		"INSERT INTO finance_client VALUES (2, 2, 'mary-missing', 'DEMANDED', NULL, '87654321');",

		"INSERT INTO refund VALUES (1, 99, '2019-01-05', 32000, 'APPROVED', 'A processing refund', 99, '2025-06-01 00:00:00', 99, '2025-06-02 00:00:00', '2026-06-03 00:00:00')",
		"INSERT INTO refund VALUES (2, 2, '2019-01-06', 15500, 'APPROVED', 'An approved refund', 99, '2025-06-01 00:00:00', 99, '2025-06-02 00:00:00')",

		"INSERT INTO bank_details VALUES (1, 1, 'MR IAN TEST', '11111111', '11-11-11');",
		"INSERT INTO bank_details VALUES (2, 2, 'MS MARY MISSING', '11111111', '11-11-11');",
	)

	dispatch := &mockDispatch{}
	s := NewService(seeder.Conn, dispatch, nil, nil, nil)

	records := [][]string{
		{"Court reference", "Amount", "Bank account name", "Bank account number", "Bank account sort code", "Created by", "Approved by"},
		{"12345678", "320.00", "MR IAN TEST", "11111111", "111111", "Felicity Finance", "Morty Manager"},     // success
		{"12345678", "320.00", "MR IAN TEST", "11111111", "111111", "Felicity Finance", "Morty Manager"},     // fail - duplicate
		{"87654321", "155.00", "MS MARY MISSING", "11111111", "111111", "Felicity Finance", "Morty Manager"}, // fail - missing (Refund not set to processing)
	}

	expectedFailedLines := map[int]string{
		2: "REFUND_NOT_FOUND_OR_DUPLICATE",
		3: "REFUND_NOT_FOUND_OR_DUPLICATE",
	}

	suite.T().Run("ProcessFulfilledPayments", func(t *testing.T) {
		var currentLedgerId int
		_ = seeder.QueryRow(suite.ctx, `SELECT MAX(id) FROM ledger`).Scan(&currentLedgerId)

		var failedLines map[int]string
		failedLines, err := s.ProcessFulfilledRefunds(suite.ctx, records, shared.NewDate("2024-01-01"))
		assert.NoError(t, err)
		assert.Equal(t, expectedFailedLines, failedLines)

		var createdLedgerAllocations []createdLedgerAllocation

		rows, _ := seeder.Query(suite.ctx,
			`SELECT l.amount, l.type, l.status, l.bankdate, la.amount, la.status
						FROM ledger l
						JOIN ledger_allocation la ON l.id = la.ledger_id
					WHERE l.finance_client_id = $1 AND l.id > $2`, 99, currentLedgerId)

		for rows.Next() {
			var r createdLedgerAllocation
			_ = rows.Scan(&r.ledgerAmount, &r.ledgerType, &r.ledgerStatus, &r.datetime, &r.allocationAmount, &r.allocationStatus)
			createdLedgerAllocations = append(createdLedgerAllocations, r)
		}

		bankDate, _ := time.Parse("2006-01-02", "2024-01-01")

		assert.Equal(t, 1, len(createdLedgerAllocations))
		assert.Equal(t, createdLedgerAllocation{
			ledgerAmount:     32000,
			ledgerType:       shared.TransactionTypeRefund.Key(),
			ledgerStatus:     "CONFIRMED",
			datetime:         bankDate,
			allocationAmount: 32000,
			allocationStatus: "REAPPLIED",
		}, createdLedgerAllocations[0])

		_ = seeder.QueryRow(suite.ctx, `SELECT MAX(id) FROM ledger`).Scan(&currentLedgerId)

		var fulfilledAt time.Time
		_ = seeder.QueryRow(suite.ctx, `SELECT fulfilled_at FROM refund WHERE id = 1`).Scan(&fulfilledAt)

		assert.NotEqual(t, fulfilledAt, time.Time{})
	})
}

func Test_getRefundDetails(t *testing.T) {
	ctx := auth.Context{
		Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test")),
		User:    &shared.User{ID: 10},
	}
	record := []string{"12345678", "320.00", "MR IAN TEST", "11111111", "11-11-11", "Felicity Finance", "Morty Manager"}
	refundDetails := getRefundDetails(ctx, record, shared.NewDate("01/01/2025"), 1, &map[int]string{})
	expected := shared.FulfilledRefundDetails{
		CourtRef:      pgtype.Text{String: "12345678", Valid: true},
		Amount:        pgtype.Int4{Int32: 32000, Valid: true},
		AccountName:   pgtype.Text{String: "MR IAN TEST", Valid: true},
		AccountNumber: pgtype.Text{String: "11111111", Valid: true},
		SortCode:      pgtype.Text{String: "11-11-11", Valid: true},
		UploadedBy:    pgtype.Int4{Int32: 10, Valid: true},
		BankDate:      pgtype.Date{Time: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true},
	}
	assert.Equal(t, expected, refundDetails)
}
