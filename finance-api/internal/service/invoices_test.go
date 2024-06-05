package service

import (
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func (suite *IntegrationSuite) TestService_GetInvoices() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (7, 1, '1234', 'DEMANDED', NULL);",
		"INSERT INTO finance_client VALUES (3, 2, '1234', 'DEMANDED', NULL);",
		"INSERT INTO fee_reduction VALUES (2, 7, 'REMISSION', NULL, '2019-04-01'::DATE, '2020-03-31'::DATE, 'notes', FALSE, '2019-05-01'::DATE);",
		"INSERT INTO invoice VALUES (1, 1, 7, 'S2', 'S203531/19', '2019-04-01', '2020-03-31', 12300, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, 12300, '2019-06-06', 99);",
		"INSERT INTO ledger VALUES (1, 'random1223', '2022-04-11T08:36:40+00:00', '', 12300, '', 'Card Payment', 'APPROVED', 7, 1, 2, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);",
		"INSERT INTO ledger_allocation VALUES (1, 1, 1, '2022-04-11T08:36:40+00:00', 12300, 'APPROVED', NULL, 'Notes here', '2022-04-11', NULL);",
		"INSERT INTO invoice_fee_range VALUES (1, 1, 'General', '2022-04-01', '2023-03-31', 12300);",
	)

	Store := store.New(conn)
	dateString := "2020-03-16"
	date, _ := time.Parse("2006-01-02", dateString)
	tests := []struct {
		name    string
		id      int
		want    *shared.Invoices
		wantErr bool
	}{
		{
			name: "returns invoices when clientId matches clientId in invoice table",
			id:   1,
			want: &shared.Invoices{
				shared.Invoice{
					Id:                 1,
					Ref:                "S203531/19",
					Status:             "",
					Amount:             12300,
					RaisedDate:         shared.Date{Time: date},
					Received:           12300,
					OutstandingBalance: 0,
					Ledgers: []shared.Ledger{
						{
							Amount:          12300,
							ReceivedDate:    shared.NewDate("04/12/2022"),
							TransactionType: "Card Payment",
							Status:          "APPROVED",
						},
					},
					SupervisionLevels: []shared.SupervisionLevel{
						{
							Level:  "General",
							Amount: 12300,
							From:   shared.NewDate("01/04/2022"),
							To:     shared.NewDate("31/03/2023"),
						},
					},
				},
			},
		},
		{
			name: "returns an empty array when no match is found",
			id:   2,
			want: &shared.Invoices{},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			s := &Service{
				store: Store,
			}
			got, err := s.GetInvoices(tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetInvoices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err == nil) && len(*tt.want) == 0 {
				assert.Empty(t, got)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInvoices() got = %v, want %v", got, tt.want)
			}
		})
	}
}
