package service

import (
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
		"INSERT INTO ledger VALUES (NEXTVAL('ledger_id_seq'), 'existing', '2022-04-11T00:00:00+00:00', '', 10300, '', 'CARD PAYMENT', 'CONFIRMED', 3, NULL, NULL, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 2);",
		"INSERT INTO ledger_allocation VALUES (NEXTVAL('ledger_allocation_id_seq'), 1, 3, '2022-04-11T00:00:00+00:00', 10300, 'ALLOCATED', NULL, 'Notes here', '2022-04-11', NULL);",
		"INSERT INTO invoice_adjustment VALUES (NEXTVAL('invoice_adjustment_id_seq'), 3, 3, '2024-01-01', 'CREDIT MEMO', '5000', 'approve me', 'PENDING', '2024-01-01', 1)",

		"INSERT INTO finance_client VALUES (4, 4, '1234', 'DEMANDED', NULL);",
		"INSERT INTO invoice VALUES (4, 4, 4, 'S2', 'paid', '2019-04-01', '2020-03-31', 12300, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, NULL, '2019-06-06', NULL);",
		"INSERT INTO invoice VALUES (5, 4, 4, 'S2', 'unpaid', '2019-04-01', '2020-03-31', 12300, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, NULL, '2019-06-06', NULL);",
		"INSERT INTO ledger VALUES (NEXTVAL('ledger_id_seq'), 'fully-paid', '2022-04-11T00:00:00+00:00', '', 12300, '', 'CARD PAYMENT', 'CONFIRMED', 4, NULL, NULL, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 2);",
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

func (suite *IntegrationSuite) TestService_UpdatePendingInvoiceAdjustmentHandlesLedgersWhichAreNotConfirmed() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, '1234', 'DEMANDED', NULL);",
		"INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'reject', '2019-04-01', '2020-03-31', 12300, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, NULL, '2019-06-06', NULL);",
		"INSERT INTO ledger VALUES (NEXTVAL('ledger_id_seq'), 'existing', '2022-04-11T00:00:00+00:00', '', 10300, '', 'CARD PAYMENT', 'CONFIRMED', 1, NULL, NULL, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 2);",
		"INSERT INTO invoice_adjustment VALUES (NEXTVAL('invoice_adjustment_id_seq'), 1, 1, '2024-01-01', 'CREDIT MEMO', '5000', 'reject me', 'PENDING', '2024-01-01', CURRVAL('ledger_id_seq'))",
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
				amount   int
			}
			q := seeder.QueryRow(suite.ctx, "SELECT status, ledger_id, amount FROM invoice_adjustment WHERE id = $1", tt.args.adjustmentId)
			_ = q.Scan(
				&adjusted.status,
				&adjusted.ledgerId,
				&adjusted.amount,
			)
			assert.Equal(t, tt.args.status.Key(), adjusted.status)
			//check the invoices arent being returned somehow
			assert.Equal(t, 5000, adjusted.amount)
		})
	}
}

//
//func (suite *IntegrationSuite) Test_setAdjustmentDecision_NotLinkedToConfirmedLedgerReturnsNull() {
//	{
//		ctx := suite.ctx
//		seeder := suite.cm.Seeder(ctx, suite.T())
//
//		seeder.SeedData(
//			"INSERT INTO finance_client VALUES (1, 1, '1234', 'DEMANDED', NULL);",
//			"INSERT INTO invoice VALUES (8, 1, 1, 'S2', 'S204642/19', '2022-04-02', '2022-04-02', 11000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",
//			"INSERT INTO ledger VALUES (3, 'abc1', '2022-04-02T00:00:00+00:00', 22000, 200, 'Initial payment', 'UNKNOWN DEBIT', 'CONFIRMED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
//			"INSERT INTO invoice_adjustment VALUES (1, 1, 8, '2022-04-02', 'FEE REDUCTION REVERSAL', 5000, 'test fee reduction reversal of all fee reductions', 'APPROVED', '2022-04-02 00:00:00', 3);",
//		)
//
//		s := NewService(seeder.Conn, nil, nil, nil, nil)
//		suite.T().Run("test", func(t *testing.T) {
//			tx, _ := s.BeginStoreTx(ctx)
//			var updatedBy pgtype.Int4
//			_ = store.ToInt4(&updatedBy, ctx.(auth.Context).User.ID)
//
//			adjustment, returnedError := tx.SetAdjustmentDecision(ctx, store.SetAdjustmentDecisionParams{
//				ID: 8, Status: "ALLOCATED", UpdatedBy: updatedBy,
//			})
//			fmt.Println("Adjustment")
//			fmt.Println(adjustment.Amount)
//			fmt.Println(adjustment.AdjustmentType)
//			fmt.Println(adjustment.FinanceClientID)
//			fmt.Println(adjustment.Outstanding)
//
//			if returnedError != nil {
//				assert.Equal(t, 0, adjustment.Amount)
//				assert.Equal(t, 0, adjustment.Outstanding)
//				assert.Equal(t, 8, adjustment.InvoiceID)
//			}
//		})
//	}
//}

//func (suite *IntegrationSuite) Test_setAdjustmentDecision_LinkedToConfirmedLedgerReturnsCorrectly() {
//	{
//		ctx := suite.ctx
//		seeder := suite.cm.Seeder(ctx, suite.T())
//
//		seeder.SeedData(
//			"INSERT INTO finance_client VALUES (1, 1, '1234', 'DEMANDED', NULL);",
//			"INSERT INTO invoice VALUES (8, 1, 1, 'S2', 'S204642/19', '2022-04-02', '2022-04-02', 11000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",
//			"INSERT INTO ledger VALUES (3, 'abc1', '2022-04-02T00:00:00+00:00', 22000, 200, 'Initial payment', 'UNKNOWN DEBIT', 'CONFIRMED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
//			"INSERT INTO invoice_adjustment VALUES (1, 1, 8, '2022-04-02', 'FEE REDUCTION REVERSAL', 5000, 'test fee reduction reversal of all fee reductions', 'CONFIRMED', '2022-04-02 00:00:00', 3);",
//		)
//
//		s := NewService(seeder.Conn, nil, nil, nil, nil)
//		suite.T().Run("test", func(t *testing.T) {
//			tx, _ := s.BeginStoreTx(ctx)
//			var updatedBy pgtype.Int4
//			_ = store.ToInt4(&updatedBy, ctx.(auth.Context).User.ID)
//
//			adjustment, returnedError := tx.SetAdjustmentDecision(ctx, store.SetAdjustmentDecisionParams{
//				ID: 8, Status: "ALLOCATED", UpdatedBy: updatedBy,
//			})
//			fmt.Println(adjustment)
//
//			var (
//				amount int
//				id     int
//			)
//
//			if returnedError == nil {
//				rows := seeder.QueryRow(ctx, "SELECT id, amount FROM invoice_adjustment WHERE id = $1)", 1)
//				_ = rows.Scan(&id, &amount)
//				assert.Equal(suite.T(), 1, id)
//				assert.Equal(suite.T(), 1234, amount)
//			}
//		})
//	}
//}
