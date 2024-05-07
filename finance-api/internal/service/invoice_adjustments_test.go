package service

import (
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestService_GetInvoiceAdjustments(t *testing.T) {
	conn := testDB.GetConn()
	t.Cleanup(func() {
		testDB.Restore()
	})

	conn.SeedData(
		"INSERT INTO finance_client VALUES (6, 6, '1234', 'DEMANDED', null, 12300, 2222);",
		"INSERT INTO ledger VALUES (2, 'abc1', '2022-04-02T00:00:00+00:00', '', 12300, 'first credit', 'CREDIT MEMO', 'REJECTED', 6, null, null, null, null, null, null, null, null, '05/05/2022', 1);",
		"INSERT INTO ledger VALUES (3, 'def2', '2022-04-03T00:00:00+00:00', '', 23001, 'first write off', 'CREDIT WRITE OFF', 'CONFIRMED', 6, null, null, null, null, null, null, null, null, '05/05/2022', 1);",
		"INSERT INTO ledger VALUES (4, 'ghi3', '2022-04-04T00:00:00+00:00', '', 30023, 'second credit', 'CREDIT MEMO', 'PENDING', 6, null, null, null, null, null, null, null, null, '05/05/2022', 1);",

		// two invoices
		"INSERT INTO invoice VALUES (2, 1, 6, 'S2', 'S204642/19', '2022-04-02', '2022-04-02', 0, null, null, null, null, null, null, 0, '2022-04-02', 1);",
		"INSERT INTO invoice VALUES (3, 1, 6, 'S2', 'S205753/20', '2022-04-02', '2022-04-02', 0, null, null, null, null, null, null, 0, '2022-04-02', 1);",

		// one for each ledger
		"INSERT INTO ledger_allocation VALUES (2, 2, 2, '2022-04-02T00:00:00+00:00', 0, 'PENDING', null, '', '2022-04-02', null);",
		"INSERT INTO ledger_allocation VALUES (3, 3, 2, '2022-04-02T00:00:00+00:00', 0, 'PENDING', null, '', '2022-04-02', null);",
		"INSERT INTO ledger_allocation VALUES (4, 4, 3, '2022-04-02T00:00:00+00:00', 0, 'PENDING', null, '', '2022-04-02', null);",
	)

	dateString := "2022-04-02"
	date, _ := time.Parse("2006-01-02", dateString)
	s := NewService(conn.Conn)

	tests := []struct {
		name    string
		id      int
		want    *shared.InvoiceAdjustments
		wantErr bool
	}{
		{
			name: "returns invoice adjustments when clientId matches clientId in ledgers table",
			id:   6,
			want: &shared.InvoiceAdjustments{
				shared.InvoiceAdjustment{
					Id:             4,
					InvoiceRef:     "S205753/20",
					RaisedDate:     shared.Date{Time: date.AddDate(0, 0, 2)},
					AdjustmentType: shared.AdjustmentTypeAddCredit,
					Amount:         30023,
					Status:         "PENDING",
					Notes:          "second credit",
				},
				shared.InvoiceAdjustment{
					Id:             3,
					InvoiceRef:     "S204642/19",
					RaisedDate:     shared.Date{Time: date.AddDate(0, 0, 1)},
					AdjustmentType: shared.AdjustmentTypeWriteOff,
					Amount:         23001,
					Status:         "CONFIRMED",
					Notes:          "first write off",
				},
				shared.InvoiceAdjustment{
					Id:             2,
					InvoiceRef:     "S204642/19",
					RaisedDate:     shared.Date{Time: date},
					AdjustmentType: shared.AdjustmentTypeAddCredit,
					Amount:         12300,
					Status:         "REJECTED",
					Notes:          "first credit",
				},
			},
		},
		{
			name: "returns an empty array when no match is found",
			id:   2,
			want: &shared.InvoiceAdjustments{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.GetInvoiceAdjustments(tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetInvoiceAdjustments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err == nil) && len(*tt.want) == 0 {
				assert.Empty(t, got)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInvoiceAdjustments() got = %v, want %v", got, tt.want)
			}
		})
	}
}
