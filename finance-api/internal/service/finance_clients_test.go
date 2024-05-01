package service

import (
	"context"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"reflect"
	"testing"
)

func TestService_GetAccountInformation(t *testing.T) {
	t.Cleanup(func() {
		err := testDB.Container.Restore(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	})

	testDB.SeedData("INSERT INTO finance_client VALUES (1, 2, 'sop123', 'DEMANDED', 3, 12300, 321)")

	Store := store.New(testDB.DbInstance)
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
