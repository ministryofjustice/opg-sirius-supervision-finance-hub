package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/finance-api/internal/testhelpers"
	"github.com/opg-sirius-finance-hub/shared"
	"log"
	"os"
	"reflect"
	"testing"
)

var testDbInstance *pgxpool.Pool

func TestMain(m *testing.M) {
	testDB := testhelpers.InitDb()
	testDbInstance = testDB.DbInstance
	defer testDB.TearDown()
	os.Exit(m.Run())
}

func seedData(db *pgxpool.Pool, sqlQuery string) {
	_, err := db.Exec(context.Background(), sqlQuery)
	if err != nil {
		log.Fatal("Unable to seed data with db connection")
	}
}

func TestService_GetAccountInformation(t *testing.T) {
	sqlQuery := "INSERT INTO finance_client VALUES (1, 2, 'sop123', 'DEMANDED', 3, 12300, 321)"
	seedData(testDbInstance, sqlQuery)

	Store := store.New(testDbInstance)
	tests := []struct {
		name    string
		id      int
		want    *shared.AccountInformation
		wantErr bool
	}{
		{
			name: "returns account information when clientId matches clientId in finance_client table",
			id:   2,
			want: &shared.AccountInformation{
				OutstandingBalance: 12300,
				CreditBalance:      321,
				PaymentMethod:      "DEMANDED",
			},
		},
		{
			name:    "returns error when no match is found",
			id:      1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				Store: Store,
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
