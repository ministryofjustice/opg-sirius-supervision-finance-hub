package service

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
)

func (suite *IntegrationSuite) TestService_UpdatePendingInvoiceAdjustment() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, '1234', 'DEMANDED', NULL);",
		"INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'reject', '2019-04-01', '2020-03-31', 12300, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, NULL, '2019-06-06', NULL);",
		"INSERT INTO invoice_adjustment VALUES (NEXTVAL('invoice_adjustment_id_seq'), 1, 1, '2024-01-01', 'CREDIT MEMO', '5000', 'reject me', 'PENDING', '2024-01-01', 1)",

		"INSERT INTO finance_client VALUES (2, 2, '1234', 'DEMANDED', NULL);",
		"INSERT INTO invoice VALUES (2, 2, 2, 'S2', 'approve', '2019-04-01', '2020-03-31', 12300, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, NULL, '2019-06-06', NULL);",
		"INSERT INTO invoice_adjustment VALUES (NEXTVAL('invoice_adjustment_id_seq'), 2, 2, '2024-01-01', 'CREDIT MEMO', '5000', 'approve me', 'PENDING', '2024-01-01', 1)",

		"INSERT INTO finance_client VALUES (3, 3, '1234', 'DEMANDED', NULL);",
		"INSERT INTO invoice VALUES (3, 3, 3, 'S2', 'overpaid', '2019-04-01', '2020-03-31', 12300, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, NULL, '2019-06-06', NULL);",
		"INSERT INTO ledger VALUES (NEXTVAL('ledger_id_seq'), 'existing', '2022-04-11T00:00:00+00:00', '', 10300, '', 'CARD PAYMENT', 'CONFIRMED', 3, NULL, NULL, '2022-04-11', '2022-04-12', 1254, '', '', 1, '2022-05-05', 2, NULL, '2022-05-05');",
		"INSERT INTO ledger_allocation VALUES (NEXTVAL('ledger_allocation_id_seq'), 1, 3, '2022-04-11T00:00:00+00:00', 10300, 'ALLOCATED', NULL, 'Notes here', '2022-04-11', NULL);",
		"INSERT INTO invoice_adjustment VALUES (NEXTVAL('invoice_adjustment_id_seq'), 3, 3, '2024-01-01', 'CREDIT MEMO', '5000', 'approve me', 'PENDING', '2024-01-01', 1)",

		"INSERT INTO finance_client VALUES (4, 4, '1234', 'DEMANDED', NULL);",
		"INSERT INTO invoice VALUES (4, 4, 4, 'S2', 'paid', '2019-04-01', '2020-03-31', 12300, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, NULL, '2019-06-06', NULL);",
		"INSERT INTO invoice VALUES (5, 4, 4, 'S2', 'unpaid', '2019-04-01', '2020-03-31', 12300, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, NULL, '2019-06-06', NULL);",
		"INSERT INTO ledger VALUES (NEXTVAL('ledger_id_seq'), 'fully-paid', '2022-04-11T00:00:00+00:00', '', 12300, '', 'CARD PAYMENT', 'CONFIRMED', 4, NULL, NULL, '2022-04-11', '2022-04-12', 1254, '', '', 1, '2022-05-05', 2, NULL, '2022-05-05');",
		"INSERT INTO ledger_allocation VALUES (NEXTVAL('ledger_allocation_id_seq'), 2, 4, '2022-04-11T00:00:00+00:00', 12300, 'ALLOCATED', NULL, 'Notes here', '2022-04-11', NULL);",
		"INSERT INTO invoice_adjustment VALUES (NEXTVAL('invoice_adjustment_id_seq'), 4, 4, '2024-01-01', 'CREDIT MEMO', '5000', 'approve me', 'PENDING', '2024-01-01', 1)",

		"INSERT INTO finance_client VALUES (5, 5, '1234', 'DEMANDED', NULL);",
		"INSERT INTO invoice VALUES (6, 5, 5, 'S2', 'reversal', '2019-04-01', '2020-03-31', 12300, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, NULL, '2019-06-06', NULL);",
		"INSERT INTO invoice_adjustment VALUES (NEXTVAL('invoice_adjustment_id_seq'), 5, 5, '2024-01-01', 'WRITE OFF REVERSAL', '5000', 'approve me', 'PENDING', '2024-01-01', 1)",
	)

	dispatch := &mockDispatch{}
	s := NewService(seeder.Conn, dispatch, nil, nil, nil)

	type args struct {
		clientId     int32
		adjustmentId int32
		status       shared.AdjustmentStatus
	}
	tests := []struct {
		name                       string
		args                       args
		invoiceId                  int32
		expectedAllocationStatuses []string
		expectedDebtAmount         int
	}{
		{
			name: "Rejected",
			args: args{
				clientId:     1,
				adjustmentId: 1,
				status:       shared.AdjustmentStatusRejected,
			},
			expectedAllocationStatuses: []string{},
		},
		{
			name: "Approved",
			args: args{
				clientId:     2,
				adjustmentId: 2,
				status:       shared.AdjustmentStatusApproved,
			},
			expectedAllocationStatuses: []string{"ALLOCATED"},
			expectedDebtAmount:         7300,
		},
		{
			name: "Approved - Unapply",
			args: args{
				clientId:     3,
				adjustmentId: 3,
				status:       shared.AdjustmentStatusApproved,
			},
			expectedAllocationStatuses: []string{
				"ALLOCATED", // existing allocation
				"ALLOCATED",
				"UNAPPLIED",
			},
			expectedDebtAmount: 0,
		},
		{
			name: "Approved - Reapply",
			args: args{
				clientId:     4,
				adjustmentId: 4,
				status:       shared.AdjustmentStatusApproved,
			},
			expectedAllocationStatuses: []string{
				"ALLOCATED", // existing allocation
				"ALLOCATED",
				"UNAPPLIED",
				"REAPPLIED",
			},
			expectedDebtAmount: 0,
		},
		{
			name: "Approved - Reapply",
			args: args{
				clientId:     5,
				adjustmentId: 5,
				status:       shared.AdjustmentStatusApproved,
			},
			expectedAllocationStatuses: []string{
				"ALLOCATED",
			},
			expectedDebtAmount: 0,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			err := s.UpdatePendingInvoiceAdjustment(suite.ctx, tt.args.clientId, tt.args.adjustmentId, tt.args.status)
			if err != nil {
				t.Errorf("UpdatePendingInvoiceAdjustment() error = %v", err)
				return
			}

			rows, _ := seeder.Query(suite.ctx,
				"SELECT la.status FROM ledger_allocation la JOIN ledger l ON l.id = la.ledger_id WHERE l.finance_client_id = $1", tt.args.clientId)

			statuses := []string{}
			for rows.Next() {
				var status string
				_ = rows.Scan(&status)
				statuses = append(statuses, status)
			}

			assert.EqualValues(t, tt.expectedAllocationStatuses, statuses)

			var adjusted struct {
				status   string
				ledgerId int
			}
			q := seeder.QueryRow(suite.ctx, "SELECT status, ledger_id FROM invoice_adjustment WHERE id = $1", tt.args.adjustmentId)
			_ = q.Scan(
				&adjusted.status,
				&adjusted.ledgerId,
			)
			assert.Equal(t, tt.args.status.Key(), adjusted.status)
			if tt.args.status == shared.AdjustmentStatusRejected {
				assert.Equal(t, 0, adjusted.ledgerId)
			} else {
				assert.NotEqual(t, 0, adjusted.ledgerId) // asserts ledgerId is set
			}
		})
	}
}

func (suite *IntegrationSuite) Test_setAdjustmentDecision_LinkedToNonConfirmedLedgerDoesNotApplyInvoiceReduction() {
	{
		ctx := suite.ctx
		seeder := suite.cm.Seeder(ctx, suite.T())

		seeder.SeedData(
			"INSERT INTO finance_client VALUES (1, 1, '1234', 'DEMANDED', NULL);",
			"INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'reject', '2019-04-01', '2020-03-31', 12300, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, NULL, '2019-06-06', NULL);",
			"INSERT INTO ledger VALUES (NEXTVAL('ledger_id_seq'), 'fully-paid', '2022-04-11T00:00:00+00:00', '', 12300, '', 'CARD PAYMENT', 'APPROVED', 1, NULL, NULL, '2022-04-11', '2022-04-12', 1254, '', '', 1, '2022-05-05', 2, NULL, '2022-05-05');",
			"INSERT INTO invoice_adjustment VALUES (NEXTVAL('invoice_adjustment_id_seq'), 1, 1, '2024-01-01', 'CREDIT MEMO', '5000', 'reject me', 'PENDING', '2022-04-11T00:00:00+00:00', 1, null, null, CURRVAL('ledger_id_seq'));",
			"INSERT INTO ledger_allocation VALUES (1, CURRVAL('ledger_id_seq'), 1, '2024-01-01 15:30:27', 10000, 'ALLOCATED', NULL, '', '2024-01-01', NULL)",
		)

		s := NewService(seeder.Conn, nil, nil, nil, nil)
		suite.T().Run("test", func(t *testing.T) {
			tx, _ := s.BeginStoreTx(ctx)
			var updatedBy pgtype.Int4
			_ = store.ToInt4(&updatedBy, ctx.(auth.Context).User.ID)

			adjustment, _ := tx.SetAdjustmentDecision(ctx, store.SetAdjustmentDecisionParams{
				ID: 1, Status: "ALLOCATED", UpdatedBy: updatedBy,
			})

			assert.Equal(suite.T(), int32(1), adjustment.InvoiceID)
			assert.Equal(suite.T(), int32(12300), adjustment.Outstanding)
			assert.Equal(suite.T(), int32(1), adjustment.FinanceClientID)
			assert.Equal(suite.T(), int32(5000), adjustment.Amount)
		})
	}
}

func (suite *IntegrationSuite) Test_setAdjustmentDecision_LinkedToConfirmedLedgerAppliesInvoiceReduction() {
	{
		ctx := suite.ctx
		seeder := suite.cm.Seeder(ctx, suite.T())

		seeder.SeedData(
			"INSERT INTO finance_client VALUES (1, 1, '1234', 'DEMANDED', NULL);",
			"INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'reject', '2019-04-01', '2020-03-31', 12300, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, NULL, '2019-06-06', NULL);",
			"INSERT INTO ledger VALUES (NEXTVAL('ledger_id_seq'), 'fully-paid', '2022-04-11T00:00:00+00:00', '', 12300, '', 'CARD PAYMENT', 'CONFIRMED', 1, NULL, NULL, '2022-04-11', '2022-04-12', 1254, '', '', 1, '2022-05-05', 2, NULL, '2022-05-05');",
			"INSERT INTO invoice_adjustment VALUES (NEXTVAL('invoice_adjustment_id_seq'), 1, 1, '2024-01-01', 'CREDIT MEMO', '5000', 'reject me', 'CONFIRMED', '2022-04-11T00:00:00+00:00', 1, null, null, CURRVAL('ledger_id_seq'));",
			"INSERT INTO ledger_allocation VALUES (1, CURRVAL('ledger_id_seq'), 1, '2024-01-01 15:30:27', 10000, 'ALLOCATED', NULL, '', '2024-01-01', NULL)",
		)

		s := NewService(seeder.Conn, nil, nil, nil, nil)
		suite.T().Run("test", func(t *testing.T) {
			tx, _ := s.BeginStoreTx(ctx)
			var updatedBy pgtype.Int4
			_ = store.ToInt4(&updatedBy, ctx.(auth.Context).User.ID)

			adjustment, _ := tx.SetAdjustmentDecision(ctx, store.SetAdjustmentDecisionParams{
				ID: 1, Status: "ALLOCATED", UpdatedBy: updatedBy,
			})

			assert.Equal(suite.T(), int32(1), adjustment.InvoiceID)
			assert.Equal(suite.T(), int32(2300), adjustment.Outstanding)
			assert.Equal(suite.T(), int32(1), adjustment.FinanceClientID)
			assert.Equal(suite.T(), int32(5000), adjustment.Amount)
		})
	}
}
