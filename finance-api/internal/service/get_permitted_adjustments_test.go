package service

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
)

func (suite *IntegrationSuite) TestService_GetPermittedAdjustments() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, 'sop123', 'DEMANDED', 3)",
		// two invoices
		"INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'S204642/19', '2022-04-02', '2022-04-02', 32000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",
		"INSERT INTO invoice VALUES (2, 1, 1, 'S2', 'S204643/20', '2022-04-02', '2022-04-02', 32000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",
		"INSERT INTO invoice VALUES (3, 1, 1, 'AD', 'AD05754/20', '2022-04-02', '2022-04-02', 10000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",
		"INSERT INTO invoice VALUES (4, 1, 1, 'S2', 'AD05755/20', '2022-04-02', '2022-04-02', 32000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",
		"INSERT INTO invoice VALUES (5, 1, 1, 'AD', 'AD05756/20', '2022-04-02', '2022-04-02', 10000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",
		"INSERT INTO invoice VALUES (6, 1, 1, 'S2', 'AD05757/20', '2022-04-02', '2022-04-02', 32000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",

		"INSERT INTO ledger VALUES (1, 'abc1', '2022-04-02T00:00:00+00:00', '', 32000, 'Write off', 'CREDIT WRITE OFF', 'CONFIRMED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
		"INSERT INTO ledger VALUES (2, 'abc2', '2022-04-02T00:00:00+00:00', '', 32000, 'Paid off', 'CARD PAYMENT', 'CONFIRMED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
		"INSERT INTO ledger VALUES (3, 'abc3', '2022-04-03T00:00:00+00:00', '', 1, 'deposit', 'CARD PAYMENT', 'CONFIRMED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
		"INSERT INTO ledger VALUES (4, 'abc4', '2022-04-04T00:00:00+00:00', '', 1, 'deposit', 'CARD PAYMENT', 'CONFIRMED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",

		// one for each ledger
		"INSERT INTO ledger_allocation VALUES (1, 1, 1, '2022-04-02T00:00:00+00:00', 32000, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",
		"INSERT INTO ledger_allocation VALUES (2, 2, 2, '2022-04-02T00:00:00+00:00', 32000, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",
		"INSERT INTO ledger_allocation VALUES (3, 3, 5, '2022-04-02T00:00:00+00:00', 1, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",
		"INSERT INTO ledger_allocation VALUES (4, 4, 6, '2022-04-02T00:00:00+00:00', 1, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",

		// transactions to ignore
		"INSERT INTO ledger VALUES (5, 'abc5', '2022-04-02T00:00:00+00:00', '', 32000, 'Write off', 'CREDIT WRITE OFF', 'APPROVED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
		"INSERT INTO ledger_allocation VALUES (5, 4, 2, '2022-04-02T00:00:00+00:00', 32000, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",

		// fee reduction reversal clients
		"INSERT INTO finance_client VALUES (2, 2, 'client2sop', 'DEMANDED', 4)",
		"INSERT INTO fee_reduction VALUES (1, 2, 'REMISSION', NULL, '2022-04-02', '2025-04-02', 'thou art remissed', false, '2022-04-02', '2022-04-02T00:00:00+00:00', 1);",

		"INSERT INTO invoice VALUES (7, 2, 2, 'AD', 'AD05758/20', '2022-04-02', '2022-04-02', 10000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",
		"INSERT INTO invoice VALUES (8, 2, 2, 'AD', 'AD05759/20', '2022-04-02', '2022-04-02', 10000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",

		"INSERT INTO ledger VALUES (6, 'abc6', '2022-04-02T00:00:00+00:00', '', 5000, 'oops!', 'REMISSION', 'CONFIRMED', 2, NULL, 1, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",

		"INSERT INTO ledger_allocation VALUES (6, 6, 7, '2022-04-02T00:00:00+00:00', 5000, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",
		"INSERT INTO ledger_allocation VALUES (7, 6, 8, '2022-04-02T00:00:00+00:00', 5000, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",

		"INSERT INTO invoice_adjustment VALUES (1, 2, 8, '2022-04-02', 'FEE REDUCTION REVERSAL', 5000, 'test fee reduction reversal of all fee reductions', 'APPROVED', '2022-04-02 00:00:00', 1)",

		"INSERT INTO finance_client VALUES (3, 3, 'client3sop', 'DEMANDED', 5)",
		"INSERT INTO fee_reduction VALUES (2, 3, 'EXEMPTION', NULL, '2022-04-02', '2025-04-02', 'thou art exempt', false, '2022-04-02', '2022-04-02T00:00:00+00:00', 1);",
		"INSERT INTO fee_reduction VALUES (3, 3, 'HARDSHIP', NULL, '2022-04-02', '2025-04-02', 'thou art hardshipped', false, '2022-04-02', '2022-04-02T00:00:00+00:00', 1);",

		"INSERT INTO invoice VALUES (9, 2, 2, 'AD', 'AD05760/20', '2022-04-02', '2022-04-02', 10000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",

		"INSERT INTO ledger VALUES (7, 'abc7', '2022-04-02T00:00:00+00:00', '', 5000, 'oops!', 'REMISSION', 'CONFIRMED', 3, NULL, 2, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
		"INSERT INTO ledger VALUES (8, 'abc8', '2022-04-02T00:00:00+00:00', '', 5000, 'oops!', 'REMISSION', 'CONFIRMED', 3, NULL, 3, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",

		"INSERT INTO ledger_allocation VALUES (8, 7, 9, '2022-04-02T00:00:00+00:00', 5000, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",
		"INSERT INTO ledger_allocation VALUES (9, 8, 9, '2022-04-02T00:00:00+00:00', 5000, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",

		"INSERT INTO invoice_adjustment VALUES (3, 3, 9, '2022-04-02', 'FEE REDUCTION REVERSAL', 5000, 'test fee reduction reversal of all fee reductions', 'APPROVED', '2022-04-02 00:00:00', 1)",

		"INSERT INTO invoice VALUES (10, 3, 3, 'AD', 'AD05761/20', '2022-04-02', '2022-04-02', 10000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",

		"INSERT INTO ledger VALUES (9, 'abc9', '2022-04-02T00:00:00+00:00', '', 10000, 'Write off', 'CREDIT WRITE OFF', 'CONFIRMED', 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
		"INSERT INTO ledger_allocation VALUES (10, 9, 10, '2022-04-02T00:00:00+00:00', 10000, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",

		"INSERT INTO ledger VALUES (10, 'abc10', '2022-04-02T00:00:00+00:00', '', -10000, 'Write off reversal', 'WRITE OFF REVERSAL', 'CONFIRMED', 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
		"INSERT INTO ledger_allocation VALUES (11, 10, 10, '2022-04-02T00:00:00+00:00', -10000, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",

		"INSERT INTO invoice VALUES (11, 3, 3, 'AD', 'AD05762/20', '2022-04-02', '2022-04-02', 10000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",

		"INSERT INTO ledger VALUES (11, 'abc11', '2022-04-02T00:00:00+00:00', '', 10000, 'Write off', 'CREDIT WRITE OFF', 'CONFIRMED', 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
		"INSERT INTO ledger_allocation VALUES (12, 11, 11, '2022-04-02T00:00:00+00:00', 10000, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",

		"INSERT INTO ledger VALUES (12, 'abc12', '2022-04-02T00:00:00+00:00', '', -5000, 'Write off reversal', 'WRITE OFF REVERSAL', 'CONFIRMED', 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
		"INSERT INTO ledger_allocation VALUES (13, 11, 11, '2022-04-02T00:00:00+00:00', -5000, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",
	)

	Store := store.New(seeder.Conn)
	tests := []struct {
		name    string
		id      int32
		want    []shared.AdjustmentType
		wantErr bool
	}{
		{
			name: "written off",
			id:   1,
			want: []shared.AdjustmentType{
				shared.AdjustmentTypeWriteOffReversal,
			},
		},
		{
			name: "zero balance",
			id:   2,
			want: []shared.AdjustmentType{
				shared.AdjustmentTypeCreditMemo,
				shared.AdjustmentTypeDebitMemo,
			},
		},
		{
			name: "AD full balance",
			id:   3,
			want: []shared.AdjustmentType{
				shared.AdjustmentTypeWriteOff,
				shared.AdjustmentTypeCreditMemo,
			},
		},
		{
			name: "non-AD full balance",
			id:   4,
			want: []shared.AdjustmentType{
				shared.AdjustmentTypeWriteOff,
				shared.AdjustmentTypeCreditMemo,
			},
		},
		{
			name: "AD partially paid",
			id:   5,
			want: []shared.AdjustmentType{
				shared.AdjustmentTypeWriteOff,
				shared.AdjustmentTypeCreditMemo,
				shared.AdjustmentTypeDebitMemo,
			},
		},
		{
			name: "non-AD partially paid",
			id:   6,
			want: []shared.AdjustmentType{
				shared.AdjustmentTypeWriteOff,
				shared.AdjustmentTypeCreditMemo,
				shared.AdjustmentTypeDebitMemo,
			},
		},
		{
			name:    "returns error when no match is found",
			id:      99,
			wantErr: true,
		},
		{
			name: "fee reduction and no existing reversal",
			id:   7,
			want: []shared.AdjustmentType{
				shared.AdjustmentTypeWriteOff,
				shared.AdjustmentTypeCreditMemo,
				shared.AdjustmentTypeDebitMemo,
				shared.AdjustmentTypeFeeReductionReversal,
			},
		},
		{
			name: "fee reduction fully reversed",
			id:   8,
			want: []shared.AdjustmentType{
				shared.AdjustmentTypeWriteOff,
				shared.AdjustmentTypeCreditMemo,
				shared.AdjustmentTypeDebitMemo,
			},
		},
		{
			name: "2 fee reductions, only one reversed",
			id:   9,
			want: []shared.AdjustmentType{
				shared.AdjustmentTypeCreditMemo,
				shared.AdjustmentTypeDebitMemo,
				shared.AdjustmentTypeFeeReductionReversal,
			},
		},
		{
			name: "write off fully reversed",
			id:   10,
			want: []shared.AdjustmentType{
				shared.AdjustmentTypeWriteOff,
				shared.AdjustmentTypeCreditMemo,
			},
		},
		{
			name: "write off partially reversed",
			id:   11,
			want: []shared.AdjustmentType{
				shared.AdjustmentTypeWriteOffReversal,
			},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			s := &Service{
				store: Store,
			}
			got, err := s.GetPermittedAdjustments(suite.ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPermittedAdjustments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.ElementsMatch(t, got, tt.want)
		})
	}
}
