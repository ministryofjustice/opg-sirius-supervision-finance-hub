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
				AdjustmentType:  shared.AdjustmentTypeAddCredit,
				AdjustmentNotes: "credit",
				Amount:          52000,
			},
			err: shared.BadRequest{Field: "Amount", Reason: "Amount entered must be equal to or less than £420"},
		},
		{
			name:      "Invalid client id",
			invoiceId: 1,
			clientId:  99,
			data: &shared.CreateLedgerEntryRequest{
				AdjustmentType:  shared.AdjustmentTypeAddCredit,
				AdjustmentNotes: "credit",
				Amount:          42000,
			},
			err: pgx.ErrNoRows,
		},
		{
			name:      "Ledger entry created",
			invoiceId: 1,
			clientId:  1,
			data: &shared.CreateLedgerEntryRequest{
				AdjustmentType:  shared.AdjustmentTypeAddCredit,
				AdjustmentNotes: "credit",
				Amount:          32000,
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
					Notes:           pgtype.Text{String: tt.data.AdjustmentNotes, Valid: true},
					Type:            tt.data.AdjustmentType.Key(),
					Status:          "PENDING",
					FinanceClientID: pgtype.Int4{Int32: int32(1), Valid: true},
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
					LedgerID:  pgtype.Int4{Int32: int32(2), Valid: true},
					InvoiceID: pgtype.Int4{Int32: int32(1), Valid: true},
					Amount:    int32(tt.data.Amount),
					Status:    "PENDING",
					Notes:     pgtype.Text{String: tt.data.AdjustmentNotes, Valid: true},
				}

				assert.EqualValues(t, expected, la)
			}
		})
	}
}

func TestService_ValidateAdjustmentAmount(t *testing.T) {
	s := Service{}

	testCases := []struct {
		name       string
		adjustment *shared.CreateLedgerEntryRequest
		balance    store.GetInvoiceBalanceRow
		err        error
	}{
		{
			name: "Unimplemented adjustment type",
			adjustment: &shared.CreateLedgerEntryRequest{
				AdjustmentType: shared.AdjustmentTypeUnknown,
				Amount:         52000,
			},
			balance: store.GetInvoiceBalanceRow{
				Initial:     32000,
				Outstanding: 10000,
			},
			err: shared.BadRequest{Field: "AdjustmentType", Reason: "Unimplemented adjustment type"},
		},
		{
			name: "Add Credit - too high",
			adjustment: &shared.CreateLedgerEntryRequest{
				AdjustmentType: shared.AdjustmentTypeAddCredit,
				Amount:         52000,
			},
			balance: store.GetInvoiceBalanceRow{
				Initial:     32000,
				Outstanding: 10000,
			},
			err: shared.BadRequest{Field: "Amount", Reason: "Amount entered must be equal to or less than £420"},
		},
		{
			name: "Add Credit - valid",
			adjustment: &shared.CreateLedgerEntryRequest{
				AdjustmentType: shared.AdjustmentTypeAddCredit,
				Amount:         42000,
			},
			balance: store.GetInvoiceBalanceRow{
				Initial:     32000,
				Outstanding: 10000,
			},
			err: nil,
		},
		{
			name: "Write off - no outstanding balance",
			adjustment: &shared.CreateLedgerEntryRequest{
				AdjustmentType: shared.AdjustmentTypeWriteOff,
			},
			balance: store.GetInvoiceBalanceRow{
				Initial:     32000,
				Outstanding: 0,
			},
			err: shared.BadRequest{Field: "Amount", Reason: "No outstanding balance to write off"},
		},
		{
			name: "Write off - valid",
			adjustment: &shared.CreateLedgerEntryRequest{
				AdjustmentType: shared.AdjustmentTypeWriteOff,
			},
			balance: store.GetInvoiceBalanceRow{
				Initial:     32000,
				Outstanding: 10000,
			},
			err: nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := s.validateAdjustmentAmount(tt.adjustment, tt.balance)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}

func TestService_CalculateAdjustmentAmount(t *testing.T) {
	s := Service{}

	testCases := []struct {
		name       string
		adjustment *shared.CreateLedgerEntryRequest
		balance    store.GetInvoiceBalanceRow
		expected   int32
	}{
		{
			name: "Write off returns outstanding balance",
			adjustment: &shared.CreateLedgerEntryRequest{
				AdjustmentType: shared.AdjustmentTypeWriteOff,
				Amount:         52000,
			},
			balance: store.GetInvoiceBalanceRow{
				Initial:     32000,
				Outstanding: 10000,
			},
			expected: 10000,
		},
		{
			name: "Add credit returns amount",
			adjustment: &shared.CreateLedgerEntryRequest{
				AdjustmentType: shared.AdjustmentTypeAddCredit,
				Amount:         52000,
			},
			balance: store.GetInvoiceBalanceRow{
				Initial:     32000,
				Outstanding: 10000,
			},
			expected: 52000,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, s.calculateAdjustmentAmount(tt.adjustment, tt.balance))
		})
	}
}
