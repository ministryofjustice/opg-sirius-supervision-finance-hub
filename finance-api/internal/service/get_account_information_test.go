package service

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"reflect"
	"testing"
)

func (suite *IntegrationSuite) TestService_GetAccountInformation() {
	seeder := suite.cm.Seeder(suite.ctx, nil)

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, 'sop123', 'DEMANDED', NULL)",
		"INSERT INTO finance_client VALUES (3, 3, 'sop123', 'DEMANDED', NULL)",
		"INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'Paid in full', '2019-04-01', '2020-03-31', 32000, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, 0, '2019-06-06', 99);",
		"INSERT INTO invoice VALUES (2, 1, 1, 'S2', 'Paid with unapply', '2020-04-01', '2021-03-31', 32000, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, 0, '2019-06-06', 99);",
		"INSERT INTO invoice VALUES (3, 1, 1, 'S2', 'Unpaid', '2020-04-01', '2021-03-31', 27000, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, 0, '2019-06-06', 99);",
		"INSERT INTO ledger VALUES (1, 'Paid in one', '2022-04-11T08:36:40+00:00', '', 32000, '', 'ONLINE CARD PAYMENT', 'CONFIRMED', 1, NULL, NULL, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 2);",
		"INSERT INTO ledger VALUES (2, 'Paid in one but...', '2022-04-11T08:36:40+00:00', '', 32000, '', 'ONLINE CARD PAYMENT', 'CONFIRMED', 1, NULL, NULL, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 2);",
		"INSERT INTO ledger VALUES (3, '... fee reduction causes unapply', '2022-04-11T08:36:40+00:00', '', 0, '', 'CREDIT REMISSION', 'CONFIRMED', 1, NULL, NULL, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 2);",
		"INSERT INTO ledger VALUES (4, 'Refund', '2022-04-11T08:36:40+00:00', '', 5000, '', 'REFUND', 'CONFIRMED', 1, NULL, NULL, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 2);",
		"INSERT INTO ledger_allocation VALUES (1, 1, 1, '2022-04-11T08:36:40+00:00', 32000, 'ALLOCATED', NULL, 'paid 1', '2022-04-11', NULL);",
		"INSERT INTO ledger_allocation VALUES (2, 2, 2, '2022-04-11T08:36:40+00:00', 32000, 'ALLOCATED', NULL, 'paid 2', '2022-04-11', NULL);",
		"INSERT INTO ledger_allocation VALUES (3, 3, 2, '2022-04-11T08:36:40+00:00', 10000, 'ALLOCATED', NULL, 'Fee reduction', '2022-04-11', NULL);",
		"INSERT INTO ledger_allocation VALUES (4, 3, 2, '2022-04-11T08:36:40+00:00', -10000, 'UNAPPLIED', NULL, 'Unapplied allocation', '2022-04-11', NULL);",
		"INSERT INTO ledger_allocation VALUES (5, 3, NULL, '2022-04-11T08:36:40+00:00', 1000000, 'PENDING', NULL, 'Ignore me', '2022-04-11', NULL);",
		"INSERT INTO ledger_allocation VALUES (6, 4, NULL, '2022-04-11T08:36:40+00:00', 5000, 'REAPPLIED', NULL, 'Refund', '2022-04-11', NULL);",
	)

	Store := store.New(seeder)
	tests := []struct {
		name    string
		id      int32
		want    *shared.AccountInformation
		wantErr bool
	}{
		{
			name: "returns account information when clientId matches clientId in finance_client table",
			id:   1,
			want: &shared.AccountInformation{
				OutstandingBalance: 27000,
				CreditBalance:      5000,
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
			got, err := s.GetAccountInformation(suite.ctx, tt.id)
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
