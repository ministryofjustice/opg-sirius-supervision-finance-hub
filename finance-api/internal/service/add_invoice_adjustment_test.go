package service

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func (suite *IntegrationSuite) TestService_AddInvoiceAdjustment() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, '1234', 'DEMANDED', NULL);",
		"INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'S204642/19', '2022-04-02', '2022-04-02', 32000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",
		"INSERT INTO ledger VALUES (1, 'abc1', '2022-04-02T00:00:00+00:00', '', 22000, 'Initial payment', 'UNKNOWN DEBIT', 'CONFIRMED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
		"ALTER SEQUENCE ledger_id_seq RESTART WITH 2;",
		"INSERT INTO ledger_allocation VALUES (1, 1, 1, '2022-04-02T00:00:00+00:00', 22000, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",
		"ALTER SEQUENCE ledger_allocation_id_seq RESTART WITH 2;",
	)

	s := NewService(seeder.Conn, nil, nil, nil, nil, nil)

	testCases := []struct {
		name      string
		invoiceId int32
		clientId  int32
		data      *shared.AddInvoiceAdjustmentRequest
		err       error
	}{
		{
			name:      "Invalid adjustment Amount",
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
			q := seeder.QueryRow(ctx, "SELECT id, finance_client_id, invoice_id, raised_date, adjustment_type, amount, notes, status FROM invoice_adjustment LIMIT 1")
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
				FinanceClientID: tt.clientId,
				InvoiceID:       tt.invoiceId,
				RaisedDate:      pgtype.Date{Time: time.Now().UTC().Truncate(24 * time.Hour), Valid: true},
				AdjustmentType:  tt.data.AdjustmentType.Key(),
				Amount:          tt.data.Amount,
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
		name                string
		roles               []string
		adjustment          *shared.AddInvoiceAdjustmentRequest
		balance             store.GetInvoiceBalanceDetailsRow
		feeReductionDetails store.GetInvoiceFeeReductionReversalDetailsRow
		err                 error
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
			err: apierror.BadRequest{Field: "Amount", Reason: "A write off reversal cannot be added to an invoice without an associated write off"},
		},
		{
			name: "Write off reversal - valid",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeWriteOffReversal,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:        32000,
				Outstanding:    10000,
				WriteOffAmount: 1000,
			},
			err: nil,
		},
		{
			name: "Write off reversal - manager override - exceeds total write offs",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeWriteOffReversal,
				Amount:         1001,
			},
			roles: []string{"Finance Manager"},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:        32000,
				Outstanding:    10000,
				WriteOffAmount: 1000,
			},
			err: apierror.BadRequest{Field: "Amount", Reason: "The write-off reversal amount must be £10 or less"},
		},
		{
			name: "Write off reversal - manager override - valid",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeWriteOffReversal,
				Amount:         999,
			},
			roles: []string{"Finance Manager"},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:        32000,
				Outstanding:    10000,
				WriteOffAmount: 1000,
			},
			err: nil,
		},
		{
			name: "Fee reduction reversal - valid",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeFeeReductionReversal,
				Amount:         500,
			},
			feeReductionDetails: store.GetInvoiceFeeReductionReversalDetailsRow{
				ReversalTotal:     pgtype.Int8{Int64: 500, Valid: true},
				FeeReductionTotal: pgtype.Int8{Int64: 1500, Valid: true},
			},
			err: nil,
		},
		{
			name: "Fee reduction reversal - invalid amount",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeFeeReductionReversal,
				Amount:         1000,
			},
			feeReductionDetails: store.GetInvoiceFeeReductionReversalDetailsRow{
				ReversalTotal:     pgtype.Int8{Int64: 500, Valid: true},
				FeeReductionTotal: pgtype.Int8{Int64: 1000, Valid: true},
			},
			err: apierror.BadRequest{Field: "Amount", Reason: "The fee reduction reversal amount must be £5 or less"},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := auth.Context{
				Context: context.Background(),
				User: &shared.User{
					ID:          1,
					DisplayName: "Test",
					Roles:       tt.roles,
				},
			}
			err := s.validateAdjustmentAmount(ctx, tt.adjustment, tt.balance, tt.feeReductionDetails)
			if tt.err != nil {
				assert.ErrorAs(t, err, &tt.err)
				assert.Equal(t, tt.err.Error(), err.Error())
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
			name: "Add debit returns the Amount as a negative",
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
			name: "Add credit returns Amount",
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
			name: "Write off reversal returns customer credit balance if less than write off Amount",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeWriteOffReversal,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:        32000,
				Outstanding:    10000,
				WriteOffAmount: 1000,
			},
			customerCreditBalance: 500,
			expected:              -500,
		},
		{
			name: "Write off reversal returns write off Amount if less than customer credit balance",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeWriteOffReversal,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:        1000,
				Outstanding:    10000,
				WriteOffAmount: 800,
			},
			customerCreditBalance: 1500,
			expected:              -800,
		},
		{
			name: "Write off reversal with manager override returns Amount",
			adjustment: &shared.AddInvoiceAdjustmentRequest{
				AdjustmentType: shared.AdjustmentTypeWriteOffReversal,
				Amount:         4321,
			},
			balance: store.GetInvoiceBalanceDetailsRow{
				Initial:        1000,
				Outstanding:    10000,
				WriteOffAmount: 800,
			},
			customerCreditBalance: 1500,
			expected:              -4321,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, s.calculateAdjustmentAmount(tt.adjustment, tt.balance, tt.customerCreditBalance))
		})
	}
}
