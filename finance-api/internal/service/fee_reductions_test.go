package service

import (
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/finance-api/internal/testhelpers"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestService_GetFeeReductions(t *testing.T) {
	testDB := testhelpers.InitDb()
	testDbInstance = testDB.DbInstance
	defer testDB.TearDown()

	sqlQuery := "INSERT INTO finance_client VALUES (5, 3, '1234', 'DEMANDED', null, 12300, 2222);"
	seedData(testDbInstance, sqlQuery)
	sqlQueryFeeReduction := "INSERT INTO fee_reduction VALUES (1, 5, 'REMISSION', null, '2019-04-01', '2020-03-31', 'notes', false, '2019-05-01');"
	seedData(testDbInstance, sqlQueryFeeReduction)

	Store := store.New(testDbInstance)

	tests := []struct {
		name    string
		id      int
		want    *shared.FeeReductions
		wantErr bool
	}{
		{
			name: "returns invoices when clientId matches clientId in invoice table",
			id:   1,
			want: &shared.FeeReductions{
				shared.FeeReduction{
					Id:           1,
					Type:         "REMISSION",
					StartDate:    shared.NewDate("01/04/2019"),
					EndDate:      shared.NewDate("31/03/2020"),
					DateReceived: shared.NewDate("01/05/2019"),
					Status:       "Active",
					Notes:        "Remission to see the notes",
				},
			},
		},
		{
			name: "returns an empty array when no match is found",
			id:   2,
			want: &shared.FeeReductions{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				Store: Store,
			}
			got, err := s.GetFeeReductions(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFeeReductions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err == nil) && len(*got) == 0 {
				assert.Empty(t, got)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFeeReductions() got = %v, want %v", got, tt.want)
			}
		})
	}
}
