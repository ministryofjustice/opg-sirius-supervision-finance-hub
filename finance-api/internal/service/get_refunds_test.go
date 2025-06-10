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
		"INSERT INTO finance_client VALUES (2, 2, 'dontfindme', 'DEMANDED', 2)",
		"INSERT INTO refund VALUES (1, 1, '2019-01-27', '2022-04-02', 12300, 'FULFILLED', 'A fulfilled refund', 99, '2025-06-04 00:00:00', 99, '2025-06-04 00:00:00')",
		"INSERT INTO refund VALUES (2, 1, '2020-01-01', NULL, 32100, 'PENDING', 'A pending refund', 99, '2025-06-04 00:00:00', NULL, NULL)",
		"INSERT INTO refund VALUES (3, 2, '2019-01-27', '2022-04-02', 99999, 'FULFILLED', 'A refund for a different client', 99, '2025-06-04 00:00:00', 99, '2025-06-04 00:00:00')",

		"INSERT INTO bank_details VALUES (1, 2, 'Clint Client', '12345678', '11-22-33');",
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
				{
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
		{
			name: "returns an empty array when no match is found",
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
			if (err == nil) && len(tt.want) == 0 {
				assert.Empty(t, got)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Mismatch (-expected +actual):\n%s", diff)
			}
		})
	}
}
