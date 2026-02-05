package service

import (
	"reflect"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) TestService_GetInvoiceAdjustments() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, '1234', 'DEMANDED', NULL);",

		// two invoices
		"INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'S204642/19', '2022-04-02', '2022-04-02', 0, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",
		"INSERT INTO invoice VALUES (2, 1, 1, 'S2', 'S205753/20', '2022-04-02', '2022-04-02', 0, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2022-04-02', 1);",

		"INSERT INTO invoice_adjustment VALUES (2, 1, 1, '2022-04-02', 'CREDIT MEMO', 12300, 'first credit', 'REJECTED', '2022-04-02T00:00:00+00:00', 1)",
		"INSERT INTO invoice_adjustment VALUES (3, 1, 1, '2022-04-03', 'CREDIT WRITE OFF', 23001, 'first write off', 'APPROVED', '2022-04-02T00:00:00+00:00', 2)",
		"INSERT INTO invoice_adjustment VALUES (4, 1, 2, '2022-04-04', 'CREDIT MEMO', 30023, 'second credit', 'PENDING', '2022-04-02T00:00:00+00:00', 3)",
	)

	dateString := "2022-04-02"
	date, _ := time.Parse("2006-01-02", dateString)
	s := Service{store: store.New(seeder.Conn)}

	tests := []struct {
		name    string
		id      int32
		want    shared.InvoiceAdjustments
		wantErr bool
	}{
		{
			name: "returns invoice adjustments when clientId matches clientId in ledgers table",
			id:   1,
			want: shared.InvoiceAdjustments{
				shared.InvoiceAdjustment{
					Id:             4,
					InvoiceRef:     "S205753/20",
					RaisedDate:     shared.Date{Time: date.AddDate(0, 0, 2)},
					AdjustmentType: shared.AdjustmentTypeCreditMemo,
					Amount:         30023,
					Status:         "PENDING",
					Notes:          "second credit",
					CreatedBy:      3,
				},
				shared.InvoiceAdjustment{
					Id:             3,
					InvoiceRef:     "S204642/19",
					RaisedDate:     shared.Date{Time: date.AddDate(0, 0, 1)},
					AdjustmentType: shared.AdjustmentTypeWriteOff,
					Amount:         23001,
					Status:         "APPROVED",
					Notes:          "first write off",
					CreatedBy:      2,
				},
				shared.InvoiceAdjustment{
					Id:             2,
					InvoiceRef:     "S204642/19",
					RaisedDate:     shared.Date{Time: date},
					AdjustmentType: shared.AdjustmentTypeCreditMemo,
					Amount:         12300,
					Status:         "REJECTED",
					Notes:          "first credit",
					CreatedBy:      1,
				},
			},
		},
		{
			name: "returns an empty array when no match is found",
			id:   2,
			want: shared.InvoiceAdjustments{},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			got, err := s.GetInvoiceAdjustments(suite.ctx, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetInvoiceAdjustments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err == nil) && len(tt.want) == 0 {
				assert.Empty(t, got)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInvoiceAdjustments() got = %v, want %v", got, tt.want)
			}
		})
	}
}
