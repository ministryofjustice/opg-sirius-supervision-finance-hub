package service

import (
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
)

func (suite *IntegrationSuite) TestService_GetPermittedAdjustments() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, 'sop123', 'DEMANDED', 3, NULL, NULL)",
		// two invoices
		"INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'S204642/19', '2022-04-02', '2022-04-02', 32000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",
		"INSERT INTO invoice VALUES (2, 1, 1, 'S2', 'S204643/20', '2022-04-02', '2022-04-02', 32000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",
		"INSERT INTO invoice VALUES (3, 1, 1, 'AD', 'AD05754/20', '2022-04-02', '2022-04-02', 10000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",
		"INSERT INTO invoice VALUES (4, 1, 1, 'S2', 'AD05755/20', '2022-04-02', '2022-04-02', 32000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",
		"INSERT INTO invoice VALUES (5, 1, 1, 'AD', 'AD05756/20', '2022-04-02', '2022-04-02', 10000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",
		"INSERT INTO invoice VALUES (6, 1, 1, 'S2', 'AD05757/20', '2022-04-02', '2022-04-02', 32000, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",

		"INSERT INTO ledger VALUES (1, 'abc1', '2022-04-02T00:00:00+00:00', '', 32000, 'Write off', 'CREDIT WRITE OFF', 'APPROVED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
		"INSERT INTO ledger VALUES (2, 'abc2', '2022-04-02T00:00:00+00:00', '', 32000, 'Paid off', 'CARD PAYMENT', 'APPROVED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
		"INSERT INTO ledger VALUES (3, 'abc3', '2022-04-03T00:00:00+00:00', '', 1, 'deposit', 'CARD PAYMENT', 'APPROVED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
		"INSERT INTO ledger VALUES (4, 'abc4', '2022-04-04T00:00:00+00:00', '', 1, 'deposit', 'CARD PAYMENT', 'APPROVED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",

		// one for each ledger
		"INSERT INTO ledger_allocation VALUES (1, 1, 1, '2022-04-02T00:00:00+00:00', 32000, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",
		"INSERT INTO ledger_allocation VALUES (2, 2, 2, '2022-04-02T00:00:00+00:00', 32000, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",
		"INSERT INTO ledger_allocation VALUES (3, 3, 5, '2022-04-02T00:00:00+00:00', 1, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",
		"INSERT INTO ledger_allocation VALUES (4, 4, 6, '2022-04-02T00:00:00+00:00', 1, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",
	)

	Store := store.New(conn)
	tests := []struct {
		name    string
		id      int
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
				shared.AdjustmentTypeAddCredit,
				shared.AdjustmentTypeAddDebit,
			},
		},
		{
			name: "AD full balance",
			id:   3,
			want: []shared.AdjustmentType{
				shared.AdjustmentTypeWriteOff,
				shared.AdjustmentTypeAddCredit,
			},
		},
		{
			name: "non-AD full balance",
			id:   4,
			want: []shared.AdjustmentType{
				shared.AdjustmentTypeWriteOff,
				shared.AdjustmentTypeAddCredit,
			},
		},
		{
			name: "AD partially paid",
			id:   5,
			want: []shared.AdjustmentType{
				shared.AdjustmentTypeWriteOff,
				shared.AdjustmentTypeAddCredit,
				shared.AdjustmentTypeAddDebit,
			},
		},
		{
			name: "non-AD partially paid",
			id:   6,
			want: []shared.AdjustmentType{
				shared.AdjustmentTypeWriteOff,
				shared.AdjustmentTypeAddCredit,
				shared.AdjustmentTypeAddDebit,
			},
		},
		{
			name:    "returns error when no match is found",
			id:      99,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			s := &Service{
				store: Store,
			}
			got, err := s.GetPermittedAdjustments(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPermittedAdjustments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.ElementsMatch(t, got, tt.want)
		})
	}
}
