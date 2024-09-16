package service

import (
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
)

func (suite *IntegrationSuite) TestService_UpdatePendingInvoiceAdjustment() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, '1234', 'DEMANDED', NULL);",
		"INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'reject', '2019-04-01', '2020-03-31', 12300, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, 12300, '2019-06-06', NULL);",
		"INSERT INTO invoice_adjustment VALUES (NEXTVAL('invoice_adjustment_id_seq'), 1, 1, '2024-01-01', 'CREDIT MEMO', '5000', 'reject me', 'PENDING', '2024-01-01', 1)",

		"INSERT INTO finance_client VALUES (2, 2, '1234', 'DEMANDED', NULL);",
		"INSERT INTO invoice VALUES (2, 2, 2, 'S2', 'approve', '2019-04-01', '2020-03-31', 12300, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, 12300, '2019-06-06', NULL);",
		"INSERT INTO invoice_adjustment VALUES (NEXTVAL('invoice_adjustment_id_seq'), 2, 2, '2024-01-01', 'CREDIT MEMO', '5000', 'approve me', 'PENDING', '2024-01-01', 1)",

		"INSERT INTO finance_client VALUES (3, 3, '1234', 'DEMANDED', NULL);",
		"INSERT INTO invoice VALUES (3, 3, 3, 'S2', 'overpaid', '2019-04-01', '2020-03-31', 12300, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, 12300, '2019-06-06', NULL);",
		"INSERT INTO ledger VALUES (NEXTVAL('ledger_id_seq'), 'existing', '2022-04-11T00:00:00+00:00', '', 10300, '', 'CARD PAYMENT', 'CONFIRMED', 3, NULL, NULL, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);",
		"INSERT INTO ledger_allocation VALUES (NEXTVAL('ledger_allocation_id_seq'), 1, 3, '2022-04-11T00:00:00+00:00', 10300, 'ALLOCATED', NULL, 'Notes here', '2022-04-11', NULL);",
		"INSERT INTO invoice_adjustment VALUES (NEXTVAL('invoice_adjustment_id_seq'), 3, 3, '2024-01-01', 'CREDIT MEMO', '5000', 'approve me', 'PENDING', '2024-01-01', 1)",

		"INSERT INTO finance_client VALUES (4, 4, '1234', 'DEMANDED', NULL);",
		"INSERT INTO invoice VALUES (4, 4, 4, 'S2', 'paid', '2019-04-01', '2020-03-31', 12300, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, 12300, '2019-06-06', NULL);",
		"INSERT INTO invoice VALUES (5, 4, 4, 'S2', 'unpaid', '2019-04-01', '2020-03-31', 12300, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, 12300, '2019-06-06', NULL);",
		"INSERT INTO ledger VALUES (NEXTVAL('ledger_id_seq'), 'fully-paid', '2022-04-11T00:00:00+00:00', '', 12300, '', 'CARD PAYMENT', 'CONFIRMED', 4, NULL, NULL, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);",
		"INSERT INTO ledger_allocation VALUES (NEXTVAL('ledger_allocation_id_seq'), 2, 4, '2022-04-11T00:00:00+00:00', 12300, 'ALLOCATED', NULL, 'Notes here', '2022-04-11', NULL);",
		"INSERT INTO invoice_adjustment VALUES (NEXTVAL('invoice_adjustment_id_seq'), 4, 4, '2024-01-01', 'CREDIT MEMO', '5000', 'approve me', 'PENDING', '2024-01-01', 1)",
	)

	s := NewService(conn.Conn)

	type args struct {
		clientId     int
		adjustmentId int
		status       shared.AdjustmentStatus
	}
	tests := []struct {
		name                       string
		args                       args
		invoiceId                  int
		expectedAllocationStatuses []string
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
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			err := s.UpdatePendingInvoiceAdjustment(suite.ctx, tt.args.clientId, tt.args.adjustmentId, tt.args.status)
			if err != nil {
				t.Errorf("UpdatePendingInvoiceAdjustment() error = %v", err)
				return
			}

			var adjusted struct {
				status string
			}
			q := conn.QueryRow(suite.ctx, "SELECT status FROM invoice_adjustment WHERE id = $1", tt.args.adjustmentId)
			_ = q.Scan(
				&adjusted.status,
			)
			assert.Equal(t, tt.args.status.Key(), adjusted.status)

			rows, _ := conn.Query(suite.ctx,
				"SELECT la.status FROM ledger_allocation la JOIN ledger l ON l.id = la.ledger_id WHERE l.finance_client_id = $1", tt.args.clientId)

			statuses := []string{}
			for rows.Next() {
				var status string
				_ = rows.Scan(&status)
				statuses = append(statuses, status)
			}

			assert.EqualValues(t, tt.expectedAllocationStatuses, statuses)
		})
	}
}
