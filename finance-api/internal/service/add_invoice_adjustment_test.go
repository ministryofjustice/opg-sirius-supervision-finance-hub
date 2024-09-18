package service

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/apierror"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func (suite *IntegrationSuite) TestService_AddInvoiceAdjustment() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, '1234', 'DEMANDED', NULL);",
		"INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'S204642/19', '2022-04-02', '2022-04-02', 32000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",
		"INSERT INTO ledger VALUES (1, 'abc1', '2022-04-02T00:00:00+00:00', '', 22000, 'Initial payment', 'UNKNOWN DEBIT', 'CONFIRMED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
		"ALTER SEQUENCE ledger_id_seq RESTART WITH 2;",
		"INSERT INTO ledger_allocation VALUES (1, 1, 1, '2022-04-02T00:00:00+00:00', 22000, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",
		"ALTER SEQUENCE ledger_allocation_id_seq RESTART WITH 2;",
	)
	s := NewService(conn.Conn)

	testCases := []struct {
		name      string
		invoiceId int
		clientId  int
		data      *shared.AddInvoiceAdjustmentRequest
		err       error
	}{
		{
			name:      "Invalid adjustment amount",
			invoiceId: 1,
			clientId:  1,
			data: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType:  shared.AdjustmentTypeCreditMemo,
				AdjustmentNotes: "credit",
				Amount:          52000,
			},
			err: apierror.BadRequestError("Amount", "Amount entered must be equal to or less than £420", nil),
		},
		{
			name:      "Invalid client id",
			invoiceId: 1,
			clientId:  99,
			data: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType:  shared.AdjustmentTypeCreditMemo,
				AdjustmentNotes: "credit",
				Amount:          42000,
			},
			err: pgx.ErrNoRows,
		},
		{
			name:      "Ledger entry created",
			invoiceId: 1,
			clientId:  1,
			data: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType:  shared.AdjustmentTypeCreditMemo,
				AdjustmentNotes: "credit",
				Amount:          32000,
			},
			err: nil,
		},
	}

	for _, tt := range testCases {
		suite.T().Run(tt.name, func(t *testing.T) {
			ctx := suite.ctx
			_, err := s.AddInvoiceAdjustment(ctx, tt.clientId, tt.invoiceId, tt.data)
			if err != nil {
				assert.ErrorAs(t, err, &tt.err)
				return
			}

			var pendingAdjustment store.InvoiceAdjustment
			q := conn.QueryRow(ctx, "SELECT id, finance_client_id, invoice_id, raised_date, adjustment_type, amount, notes, status FROM invoice_adjustment LIMIT 1")
			_ = q.Scan(
				&pendingAdjustment.ID,
				&pendingAdjustment.FinanceClientID,
				&pendingAdjustment.InvoiceID,
				&pendingAdjustment.RaisedDate,
				&pendingAdjustment.AdjustmentType,
				&pendingAdjustment.Amount,
				&pendingAdjustment.Notes,
				&pendingAdjustment.Status,
			)

			expected := store.InvoiceAdjustment{
				ID:              1,
				FinanceClientID: int32(tt.clientId),
				InvoiceID:       int32(tt.invoiceId),
				RaisedDate:      pgtype.Date{Time: time.Now().UTC().Truncate(24 * time.Hour), Valid: true},
				AdjustmentType:  tt.data.AdjustmentType.Key(),
				Amount:          int32(tt.data.Amount),
				Notes:           tt.data.AdjustmentNotes,
				Status:          "PENDING",
			}

			assert.EqualValues(t, expected, pendingAdjustment)
		})
	}
}

func TestService_ValidateAdjustmentAmount(t *testing.T) {
	s := Service{}

	testCases := []struct {
		name       string
		adjustment *shared.AddInvoiceAdjustmentRequest
		balance    store.GetInvoiceBalanceDetailsRow
		err        error
	}{
		{
			name: "Unimplemented adjustment type",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeUnknown,
				Amount:         52000,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:     32000,
				Outstanding: 10000,
			},
			err: apierror.BadRequestError("AdjustmentType", "Unimplemented adjustment type", nil),
		},
		{
			name: "Add Credit - too high",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeCreditMemo,
				Amount:         52000,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:     32000,
				Outstanding: 10001,
			},
			err: apierror.BadRequestError("Amount", "Amount entered must be equal to or less than £420.01", nil),
		},
		{
			name: "Add Credit - valid",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeCreditMemo,
				Amount:         42000,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:     32000,
				Outstanding: 10000,
			},
			err: nil,
		},
		{
			name: "Add Debit - too high",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeDebitMemo,
				Amount:         52000,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:     32000,
				Outstanding: 10000,
				Feetype:     "S2",
			},
			err: apierror.BadRequestError("Amount", "Amount entered must be equal to or less than £220", nil),
		},
		{
			name: "Add Debit - too high (AD)",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeDebitMemo,
				Amount:         10001,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:     10000,
				Outstanding: 0,
				Feetype:     "AD",
			},
			err: apierror.BadRequestError("Amount", "Amount entered must be equal to or less than £100", nil),
		},
		{
			name: "Add Debit - valid",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeDebitMemo,
				Amount:         22000,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:     16000,
				Outstanding: 10000,
				Feetype:     "S2",
			},
			err: nil,
		},
		{
			name: "Add Debit - valid (AD)",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeDebitMemo,
				Amount:         5000,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:     10000,
				Outstanding: 5000,
				Feetype:     "AD",
			},
			err: nil,
		},
		{
			name: "Write off - no outstanding balance",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeWriteOff,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:     32000,
				Outstanding: 0,
			},
			err: apierror.BadRequestError("Amount", "No outstanding balance to write off", nil),
		},
		{
			name: "Write off - valid",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeWriteOff,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:     32000,
				Outstanding: 10000,
			},
			err: nil,
		},
		{
			name: "Write off reversal - not written off",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeWriteOffReversal,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:     32000,
				Outstanding: 10000,
			},
			err: shared.BadRequest{Field: "Amount", Reason: "A write off reversal cannot be added to an invoice without an associated write off"},
		},
		{
			name: "Write off reversal - valid",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeWriteOffReversal,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:     32000,
				Outstanding: 10000,
				WrittenOff:  true,
			},
			err: nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := s.validateAdjustmentAmount(tt.adjustment, tt.balance)
			if tt.err != nil {
				assert.ErrorAs(t, err, &tt.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_CalculateAdjustmentAmount(t *testing.T) {
	s := Service{}

	testCases := []struct {
		name                  string
		adjustment            *shared.AddInvoiceAdjustmentRequest
		balance               store.GetInvoiceBalanceDetailsRow
		customerCreditBalance int32
		writeOffAmount        int32
		expected              int32
	}{
		{
			name: "Write off returns outstanding balance",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeWriteOff,
				Amount:         52000,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:     32000,
				Outstanding: 10000,
			},
			expected: 10000,
		},
		{
			name: "Add debit returns the amount as a negative",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeDebitMemo,
				Amount:         22000,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:     32000,
				Outstanding: 10000,
			},
			expected: -22000,
		},
		{
			name: "Add credit returns amount",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeCreditMemo,
				Amount:         52000,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:     32000,
				Outstanding: 10000,
			},
			expected: 52000,
		},
		{
			name: "Write off reversal returns 0 if invoice not written off",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeWriteOffReversal,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:     32000,
				Outstanding: 10000,
			},
			expected: 0,
		},
		{
			name: "Write off reversal returns customer credit balance if less than write off amount",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeWriteOffReversal,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:     32000,
				Outstanding: 10000,
				WrittenOff:  true,
			},
			customerCreditBalance: 500,
			writeOffAmount:        1000,
			expected:              -500,
		},
		{
			name: "Write off reversal returns write off amount if less than customer credit balance",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeWriteOffReversal,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:     1000,
				Outstanding: 10000,
				WrittenOff:  true,
			},
			customerCreditBalance: 1500,
			writeOffAmount:        800,
			expected:              -800,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, s.calculateAdjustmentAmount(tt.adjustment, tt.balance, tt.customerCreditBalance, tt.writeOffAmount))
		})
	}
}
