package service

import (
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"reflect"
	"testing"
)

func (suite *IntegrationSuite) TestService_GetAccountInformation() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, 'sop123', 'DEMANDED', NULL)",
		"INSERT INTO finance_client VALUES (3, 3, 'sop123', 'DEMANDED', NULL)",
		"INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'S203531/19', '2019-04-01', '2020-03-31', 32000, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, 0, '2019-06-06', 99);",
		"INSERT INTO invoice VALUES (2, 1, 1, 'S2', 'S203532/20', '2020-04-01', '2021-03-31', 32000, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, 0, '2019-06-06', 99);",
		"INSERT INTO ledger VALUES (1, 'random123', '2022-04-11T08:36:40+00:00', '', 12000, '', 'CARD PAYMENT', 'APPROVED', 1, NULL, NULL, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);",
		"INSERT INTO ledger VALUES (2, 'random456', '2022-04-11T08:36:40+00:00', '', 12000, '', 'CARD PAYMENT', 'APPROVED', 1, NULL, NULL, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);",
		"INSERT INTO ledger_allocation VALUES (1, 1, 1, '2022-04-11T08:36:40+00:00', 12000, 'ALLOCATED', NULL, 'Notes here', '2022-04-11', NULL);",
		"INSERT INTO ledger_allocation VALUES (2, 2, 2, '2022-04-11T08:36:40+00:00', 12000, 'ALLOCATED', NULL, 'Notes here', '2022-04-11', NULL);",
	)

	Store := store.New(conn)
	tests := []struct {
		name    string
		id      int
		want    *shared.AccountInformation
		wantErr bool
	}{
		{
			name: "returns account information when clientId matches clientId in finance_client table",
			id:   1,
			want: &shared.AccountInformation{
				OutstandingBalance: 40000,
				CreditBalance:      0,
				PaymentMethod:      "DEMANDED",
			},
		},
		{
			name:    "returns error when no match is found",
			id:      2,
			wantErr: true,
		},
		{
			name: "returns payment details with zero outstanding when client exists but has no invoices",
			id:   3,
			want: &shared.AccountInformation{
				OutstandingBalance: 0,
				CreditBalance:      0,
				PaymentMethod:      "DEMANDED",
			},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			s := &Service{
				store: Store,
			}
			got, err := s.GetAccountInformation(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccountInformation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAccountInformation() got = %v, want %v", got, tt.want)
			}
		})
	}
}
