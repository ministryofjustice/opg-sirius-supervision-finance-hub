package service

import (
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

//func TestMain(m *testing.M) {
//	testDB := testhelpers.InitDb()
//	testDbInstance = testDB.DbInstance
//	defer testDB.TearDown()
//	os.Exit(m.Run())
//}
//
//func seedData(db *pgxpool.Pool, sql string) {
//	_, err := db.Exec(context.Background(), sql)
//	if err != nil {
//		log.Fatal("Unable to seed data with db connection")
//	}
//}

func TestService_GetInvoices(t *testing.T) {
	sqlQuery := "INSERT INTO finance_client VALUES (1, 1, '1234', 'DEMANDED', null, 12300, 2222);"
	seedData(testDbInstance, sqlQuery)
	sqlQueryFeeReduction := "INSERT INTO fee_reduction VALUES (1, 1, 'REMISSION', null, '2019-04-01', '2020-03-31', 'notes', false, '2019-05-01');"
	seedData(testDbInstance, sqlQueryFeeReduction)
	sqlQueryInvoice := "INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'S203531/19', '2019-04-01', '2020-03-31', 12300, null, 1, '2020-03-20', 10, '2020-03-16', null, null, 12300, '2019-06-06', 99);"
	seedData(testDbInstance, sqlQueryInvoice)
	Store := store.New(testDbInstance)
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
					Status:             "UNKNOWN",
					Amount:             12300,
					RaisedDate:         shared.Date{Time: date},
					Received:           0,
					OutstandingBalance: 12300,
					Ledgers:            nil,
					SupervisionLevels:  nil,
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
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				Store: Store,
			}
			got, err := s.GetInvoices(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInvoices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err == nil) && len(*got) == 0 {
				assert.Empty(t, got)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInvoices() got = %v, want %v", got, tt.want)
			}
		})
	}
}
