package service

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func (suite *IntegrationSuite) TestService_GetFeeReductions() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (5, 5, '1234', 'DEMANDED', NULL);",
		"INSERT INTO fee_reduction VALUES (5, 5, 'REMISSION', NULL, '2019-04-01', '2020-03-31', 'Remission to see the notes', FALSE, '2019-05-01');",
	)

	Store := store.New(seeder)

	tests := []struct {
		name    string
		id      int32
		want    shared.FeeReductions
		wantErr bool
	}{
		{
			name: "returns invoices when clientId matches clientId in invoice table",
			id:   5,
			want: shared.FeeReductions{
				shared.FeeReduction{
					Id:           5,
					Type:         shared.FeeReductionTypeRemission,
					StartDate:    shared.NewDate("01/04/2019"),
					EndDate:      shared.NewDate("31/03/2020"),
					DateReceived: shared.NewDate("01/05/2019"),
					Status:       "Expired",
					Notes:        "Remission to see the notes",
				},
			},
		},
		{
			name: "returns an empty array when no match is found",
			id:   2,
			want: shared.FeeReductions{},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			s := &Service{
				store: Store,
			}
			got, err := s.GetFeeReductions(suite.ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFeeReductions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err == nil) && len(tt.want) == 0 {
				assert.Empty(t, got)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFeeReductions() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateStatus(t *testing.T) {
	type args struct {
		startDate shared.Date
		endDate   shared.Date
		deleted   bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "returns pending when today is before start BankDate and not deleted",
			args: args{
				startDate: shared.Date{Time: time.Now().AddDate(-1, 0, 0).Truncate(time.Hour * 24)},
				endDate:   shared.Date{Time: time.Now().AddDate(1, 0, 0).Truncate(time.Hour * 24)},
				deleted:   false,
			},
			want: shared.StatusActive,
		},
		{
			name: "returns active when today is the start BankDate and before end BankDate and not deleted",
			args: args{
				startDate: shared.Date{Time: time.Now().Truncate(time.Hour * 24)},
				endDate:   shared.Date{Time: time.Now().AddDate(1, 0, 0).Truncate(time.Hour * 24)},
				deleted:   false,
			},
			want: shared.StatusActive,
		},
		{
			name: "returns active when today is after start BankDate and before end BankDate and not deleted",
			args: args{
				startDate: shared.Date{Time: time.Now().AddDate(-1, 0, 0).Truncate(time.Hour * 24)},
				endDate:   shared.Date{Time: time.Now().AddDate(1, 0, 0).Truncate(time.Hour * 24)},
				deleted:   false,
			},
			want: shared.StatusActive,
		},
		{
			name: "returns active when today is the end BankDate and not deleted",
			args: args{
				startDate: shared.Date{Time: time.Now().AddDate(-2, 0, 0).Truncate(time.Hour * 24)},
				endDate:   shared.Date{Time: time.Now().Truncate(time.Hour * 24)},
				deleted:   false,
			},
			want: shared.StatusActive,
		},
		{
			name: "returns expired when today is after end BankDate and not deleted",
			args: args{
				startDate: shared.Date{Time: time.Now().AddDate(-2, 0, 0).Truncate(time.Hour * 24)},
				endDate:   shared.Date{Time: time.Now().AddDate(-1, 0, 0).Truncate(time.Hour * 24)},
				deleted:   false,
			},
			want: "Expired",
		},
		{
			name: "returns cancelled the fee reduction is deleted",
			args: args{
				startDate: shared.Date{Time: time.Now().AddDate(-1, 0, 0).Truncate(time.Hour * 24)},
				endDate:   shared.Date{Time: time.Now().AddDate(1, 0, 0).Truncate(time.Hour * 24)},
				deleted:   true,
			},
			want: "Cancelled",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, calculateStatus(tt.args.startDate, tt.args.endDate, tt.args.deleted), "calculateStatus(%v, %v, %v)", tt.args.startDate, tt.args.endDate, tt.args.deleted)
		})
	}
}
