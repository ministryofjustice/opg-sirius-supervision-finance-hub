package service

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_CreateLedgerEntry(t *testing.T) {
	conn := testDB.GetConn()
	t.Cleanup(func() {
		testDB.Restore()
	})

	conn.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, '1234', 'DEMANDED', null, 0, 0);",
		"INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'S204642/19', '2022-04-02', '2022-04-02', 32000, null, null, null, null, null, null, 0, '2022-04-02', 1);",
		"INSERT INTO ledger VALUES (1, 'abc1', '2022-04-02T00:00:00+00:00', '', 22000, 'Initial payment', 'UNKNOWN DEBIT', 'CONFIRMED', 1, null, null, null, null, null, null, null, null, '05/05/2022', 1);",
		"ALTER SEQUENCE ledger_id_seq RESTART WITH 2;",
		"INSERT INTO ledger_allocation VALUES (1, 1, 1, '2022-04-02T00:00:00+00:00', 22000, '', null, '', '2022-04-02', null);",
		"ALTER SEQUENCE ledger_allocation_id_seq RESTART WITH 2;",
	)
	s := NewService(conn.Conn)

	testCases := []struct {
		name      string
		invoiceId int
		clientId  int
		data      *shared.CreateLedgerEntryRequest
		err       error
	}{
		{
			name:      "Invalid adjustment amount",
			invoiceId: 1,
			clientId:  1,
			data: &shared.CreateLedgerEntryRequest{
				AdjustmentType: shared.AdjustmentTypeAddCredit,
				Notes:          "credit",
				Amount:         52000,
			},
			err: shared.BadRequest{Field: "Amount", Reason: "Amount entered must be equal to or less than £420"},
		},
		{
			name:      "Invalid client id",
			invoiceId: 1,
			clientId:  99,
			data: &shared.CreateLedgerEntryRequest{
				AdjustmentType: shared.AdjustmentTypeAddCredit,
				Notes:          "credit",
				Amount:         42000,
			},
			err: pgx.ErrNoRows,
		},
		{
			name:      "Ledger entry created",
			invoiceId: 1,
			clientId:  1,
			data: &shared.CreateLedgerEntryRequest{
				AdjustmentType: shared.AdjustmentTypeAddCredit,
				Notes:          "credit",
				Amount:         32000,
			},
			err: nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := s.CreateLedgerEntry(tt.clientId, tt.invoiceId, tt.data)
			if err != nil {
				assert.ErrorIs(t, err, tt.err)
				return
			}

			var ledger store.Ledger
			q := conn.QueryRow(context.Background(), "SELECT id, amount, notes, type, status, finance_client_id FROM ledger WHERE id = 2")
			err = q.Scan(&ledger.ID, &ledger.Amount, &ledger.Notes, &ledger.Type, &ledger.Status, &ledger.FinanceClientID)
			if err != nil {
				assert.ErrorIs(t, err, tt.err)
			} else {
				expected := store.Ledger{
					ID:              2,
					Amount:          int32(tt.data.Amount),
					Notes:           pgtype.Text{tt.data.Notes, true},
					Type:            tt.data.AdjustmentType.DbValue(),
					Status:          "PENDING",
					FinanceClientID: pgtype.Int4{int32(1), true},
				}

				assert.EqualValues(t, expected, ledger)
			}

			var la store.LedgerAllocation
			q = conn.QueryRow(context.Background(), "SELECT id, ledger_id, invoice_id, amount, status, notes FROM ledger_allocation WHERE id = 2")
			err = q.Scan(&la.ID, &la.LedgerID, &la.InvoiceID, &la.Amount, &la.Status, &la.Notes)
			if err != nil {
				assert.ErrorIs(t, err, tt.err)
			} else {
				expected := store.LedgerAllocation{
					ID:        2,
					LedgerID:  pgtype.Int4{int32(2), true},
					InvoiceID: pgtype.Int4{int32(1), true},
					Amount:    int32(tt.data.Amount),
					Status:    "PENDING",
					Notes:     pgtype.Text{tt.data.Notes, true},
				}

				assert.EqualValues(t, expected, la)
			}
		})
	}
}

func TestService_ValidateAdjustmentAmount(t *testing.T) {
	conn := testDB.GetConn()
	t.Cleanup(func() {
		testDB.Restore()
	})

	conn.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, '1234', 'DEMANDED', null, 0, 0);",
		"INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'S204642/19', '2022-04-02', '2022-04-02', 32000, null, null, null, null, null, null, 0, '2022-04-02', 1);",
		"INSERT INTO ledger VALUES (1, 'abc1', '2022-04-02T00:00:00+00:00', '', 22000, 'Initial payment', 'UNKNOWN DEBIT', 'CONFIRMED', 1, null, null, null, null, null, null, null, null, '05/05/2022', 1);",
		"INSERT INTO ledger_allocation VALUES (1, 1, 1, '2022-04-02T00:00:00+00:00', 22000, '', null, '', '2022-04-02', null);",
	)

	s := NewService(conn.Conn)

	testCases := []struct {
		name      string
		invoiceId int
		data      *shared.CreateLedgerEntryRequest
		err       error
	}{
		{
			name:      "Unimplemented adjustment type",
			invoiceId: 1,
			data: &shared.CreateLedgerEntryRequest{
				AdjustmentType: shared.AdjustmentTypeUnknown,
				Amount:         52000,
			},
			err: shared.BadRequest{Field: "AdjustmentType", Reason: "Unimplemented adjustment type"},
		},
		{
			name:      "Add Credit - too high",
			invoiceId: 1,
			data: &shared.CreateLedgerEntryRequest{
				AdjustmentType: shared.AdjustmentTypeAddCredit,
				Amount:         52000,
			},
			err: shared.BadRequest{Field: "Amount", Reason: "Amount entered must be equal to or less than £420"},
		},
		{
			name:      "Add Credit - Valid",
			invoiceId: 1,
			data: &shared.CreateLedgerEntryRequest{
				AdjustmentType: shared.AdjustmentTypeAddCredit,
				Amount:         42000,
			},
			err: nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := s.validateAdjustmentAmount(tt.invoiceId, tt.data)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}
