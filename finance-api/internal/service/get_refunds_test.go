package service

import (
	"github.com/google/go-cmp/cmp"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
)

func (suite *IntegrationSuite) TestService_GetRefunds() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, 'findme', 'DEMANDED', 1)",
		"INSERT INTO ledger VALUES (1, 'abc1', '2022-04-02T00:00:00+00:00', '', 10000, 'Write off', 'CREDIT WRITE OFF', 'CONFIRMED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
		"INSERT INTO ledger_allocation VALUES (1, 1, NULL, '2022-04-02T00:00:00+00:00', -10000, 'UNAPPLIED', NULL, '', '2022-04-02', NULL);",
		"INSERT INTO refund VALUES (1, 1, '2019-01-27', '2022-04-02', 12300, 'FULFILLED', 'A fulfilled refund', 99, '2025-06-04 00:00:00', 99, '2025-06-04 00:00:00')",
		"INSERT INTO refund VALUES (2, 1, '2020-01-01', NULL, 32100, 'PENDING', 'A pending refund', 99, '2025-06-04 00:00:00', NULL, NULL)",

		"INSERT INTO bank_details VALUES (1, 2, 'Clint Client', '12345678', '11-22-33');",

		"INSERT INTO finance_client VALUES (2, 2, 'nocredit', 'DEMANDED', 2)",
		"INSERT INTO ledger VALUES (2, 'abc2', '2022-04-02T00:00:00+00:00', '', 50, 'Write off 2', 'CREDIT WRITE OFF', 'CONFIRMED', 2, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
		"INSERT INTO ledger_allocation VALUES (2, 2, NULL, '2022-04-02T00:00:00+00:00', 50, 'ALLOCATED', NULL, '', '2022-04-02', NULL);",

		"INSERT INTO finance_client VALUES (3, 3, 'norefunds', 'DEMANDED', 3)",
		"INSERT INTO ledger VALUES (3, 'abc3', '2022-04-02T00:00:00+00:00', '', 50, 'Write off 3', 'CREDIT WRITE OFF', 'CONFIRMED', 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '05/05/2022', 1);",
		"INSERT INTO ledger_allocation VALUES (3, 3, NULL, '2022-04-02T00:00:00+00:00', -50, 'UNAPPLIED', NULL, '', '2022-04-02', NULL);",

		"INSERT INTO finance_client VALUES (4, 4, 'dontfindme', 'DEMANDED', 4)",
		"INSERT INTO refund VALUES (4, 4, '2019-01-27', '2022-04-02', 99999, 'FULFILLED', 'A refund for a different client', 99, '2025-06-04 00:00:00', 99, '2025-06-04 00:00:00')",

		"INSERT INTO finance_client VALUES (99, 99, 'empty', 'DEMANDED', 99)",
	)

	s := NewService(seeder.Conn, nil, nil, nil, nil)

	fulfilledDate := "2022-04-02"

	tests := []struct {
		name    string
		id      int32
		want    shared.Refunds
		wantErr bool
	}{
		{
			name: "returns refunds by client id",
			id:   1,
			want: shared.Refunds{
				CreditBalance: 10000,
				Refunds: []shared.Refund{{
					ID:         2,
					RaisedDate: shared.NewDate("2020-01-01"),
					Amount:     32100,
					Status:     "PENDING",
					Notes:      "A pending refund",
					CreatedBy:  99,
					BankDetails: shared.NewNillable(
						&shared.BankDetails{
							Name:     "Clint Client",
							Account:  "12345678",
							SortCode: "11-22-33",
						}),
				},
					{
						ID:            1,
						RaisedDate:    shared.NewDate("2019-01-27"),
						FulfilledDate: shared.TransformNillableDate(&fulfilledDate),
						Amount:        12300,
						Status:        "FULFILLED",
						Notes:         "A fulfilled refund",
						CreatedBy:     99,
						BankDetails:   shared.Nillable[shared.BankDetails]{},
					},
				},
			},
		},
		{
			name: "Returns zero credit balance when not in credit",
			id:   2,
			want: shared.Refunds{},
		},
		{
			name: "Returns credit balance when there are no refunds",
			id:   3,
			want: shared.Refunds{
				CreditBalance: 50,
			},
		},
		{
			name: "returns an empty struct when no match is found",
			id:   99,
			want: shared.Refunds{},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			got, err := s.GetRefunds(suite.ctx, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRefunds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err == nil) && len(tt.want.Refunds) == 0 {
				assert.Empty(t, got.Refunds)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Mismatch (-expected +actual):\n%s", diff)
			}
		})
	}
}
