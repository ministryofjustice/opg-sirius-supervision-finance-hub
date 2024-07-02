package service

import (
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func Test_computeBillingHistory(t *testing.T) {
	type args struct {
		history []historyHolder
	}
	tests := []struct {
		name string
		args args
		want []shared.BillingHistory
	}{
		{
			name: "somehtg",
			args: args{history: []historyHolder{{
				billingHistory: shared.BillingHistory{
					User: "65",
					Date: shared.Date{Time: time.Date(2099, time.November, 4, 15, 4, 5, 0, time.UTC)},
					Event: shared.InvoiceAdjustmentPending{
						AdjustmentType: "credit memo",
						ClientId:       "1",
						Notes:          "credit adjustment for 12.00",
						PaymentBreakdown: shared.PaymentBreakdown{
							InvoiceReference: shared.InvoiceEvent{
								ID:        1,
								Reference: "S206666/18",
							},
							Amount: 12,
						},
						BaseBillingEvent: shared.BaseBillingEvent{Type: 6},
					},
					OutstandingBalance: 0,
				},
				balanceAdjustment: 0,
			}}},
			want: []shared.BillingHistory{{
				User: "65",
				Date: shared.Date{Time: time.Date(2099, time.November, 4, 15, 4, 5, 0, time.UTC)},
				Event: shared.InvoiceAdjustmentPending{
					AdjustmentType: "credit memo",
					ClientId:       "1",
					Notes:          "credit adjustment for 12.00",
					PaymentBreakdown: shared.PaymentBreakdown{
						InvoiceReference: shared.InvoiceEvent{
							ID:        1,
							Reference: "S206666/18",
						},
						Amount: 12,
					},
					BaseBillingEvent: shared.BaseBillingEvent{Type: 6},
				},
				OutstandingBalance: 0,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, computeBillingHistory(tt.args.history), "computeBillingHistory(%v)", tt.args.history)
		})
	}
}

func (suite *IntegrationSuite) TestService_GetBillingHistory() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (7, 1, '1234', 'DEMANDED', NULL);",
		"INSERT INTO finance_client VALUES (3, 2, '1234', 'DEMANDED', NULL);",
		"INSERT INTO invoice VALUES (1, 1, 7, 'S2', 'S203531/19', '2019-04-01', '2020-03-31', 32000, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, NULL, '2019-06-06', 99, 1);",
		"INSERT INTO ledger VALUES (1, 'random1223', '2022-04-11T08:36:40+00:00', '', 12300, '', 'CREDIT MEMO', 'PENDING', 7, 1,null, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);",
		"INSERT INTO ledger VALUES (2, 'different', '2025-04-11T08:36:40+00:00', '', 55555, '', 'DEBIT MEMO', 'PENDING', 7, 2, null, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2025', 65);",
		"INSERT INTO ledger_allocation VALUES (1, 1, 1, '2022-04-11T08:36:40+00:00', 12300, 'PENDING', NULL, 'Notes here', '2022-04-11', NULL);",
		"INSERT INTO ledger_allocation VALUES (2, 2, 1, '2022-04-11T08:36:40+00:00', 55555, 'PENDING', NULL, 'Notes here', '2022-04-11', NULL);",
	)

	Store := store.New(conn)
	dateOneString := "2022-05-05"
	dateOne, _ := time.Parse("2006-01-02", dateOneString)

	dateTwoString := "2025-05-05"
	dateTwo, _ := time.Parse("2006-01-02", dateTwoString)
	tests := []struct {
		name    string
		id      int
		want    []shared.BillingHistory
		wantErr bool
	}{
		{
			name: "returns invoices when clientId matches clientId in invoice table",
			id:   1,
			want: []shared.BillingHistory{
				{
					User: "65",
					Date: shared.Date{Time: dateTwo},
					Event: shared.InvoiceAdjustmentPending{
						AdjustmentType: "debit memo",
						ClientId:       "1",
						Notes:          "",
						PaymentBreakdown: shared.PaymentBreakdown{
							InvoiceReference: shared.InvoiceEvent{
								ID:        1,
								Reference: "S203531/19",
							},
							Amount: 555,
						},
						BaseBillingEvent: shared.BaseBillingEvent{Type: 6},
					},
					OutstandingBalance: 320,
				},
				{
					User: "65",
					Date: shared.Date{Time: dateOne},
					Event: shared.InvoiceAdjustmentPending{
						AdjustmentType: "credit memo",
						ClientId:       "1",
						Notes:          "",
						PaymentBreakdown: shared.PaymentBreakdown{
							InvoiceReference: shared.InvoiceEvent{
								ID:        1,
								Reference: "S203531/19",
							},
							Amount: 123,
						},
						BaseBillingEvent: shared.BaseBillingEvent{Type: 6},
					},
					OutstandingBalance: 320,
				},
			},
		},
		{
			name: "returns an empty array when no match is found",
			id:   2,
			want: []shared.BillingHistory{},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			s := &Service{
				store: Store,
			}
			got, err := s.GetBillingHistory(tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBillingHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err == nil) && len(tt.want) == 0 {
				assert.Empty(t, got)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBillingHistory() got = %v, want %v", got, tt.want)
			}
		})
	}
}
