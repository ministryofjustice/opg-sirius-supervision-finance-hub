package service

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func setupServiceAndParams() (*Service, shared.AddFeeReduction) {
	Store := store.New(testDB.DbInstance)
	var dateReceivedTransformed *shared.Date

	today := time.Now()
	dateInRangeOfSixMonths := today.AddDate(0, -5, -29).Format("2006-01-02")
	dateInRangeOfSixMonthsToTime, _ := time.Parse("2006-01-02", dateInRangeOfSixMonths)
	dateReceivedTransformed = &shared.Date{Time: dateInRangeOfSixMonthsToTime}

	params := shared.AddFeeReduction{
		FeeType:       "remission",
		StartYear:     "2021",
		LengthOfAward: 3,
		DateReceived:  dateReceivedTransformed,
		Notes:         "Testing",
	}

	s := &Service{
		Store: Store,
		DB:    testDB.DbConn,
	}

	return s, params
}

func TestService_AddFeeReduction(t *testing.T) {
	s, params := setupServiceAndParams()

	err := s.AddFeeReduction(22, params)
	if err == nil {
		return
	}
	t.Error("Add fee reduction failed")
}

func TestService_AddFeeReductionOverlap(t *testing.T) {
	testDB.SeedData(
		"INSERT INTO finance_client VALUES (22, 22, '1234', 'DEMANDED', null, 12300, 2222);",
		"INSERT INTO fee_reduction VALUES (22, 22, 'REMISSION', null, '2019-04-01', '2020-03-31', 'Remission to see the notes', false, '2019-05-01');",
	)
	s, params := setupServiceAndParams()
	params.StartYear = "2018"

	err := s.AddFeeReduction(22, params)
	if err != nil {
		if err.Error() == "overlap" {
			assert.Equalf(t, "overlap", err.Error(), "StartYear has an overlap")
			return
		}
	}
	t.Error("Overlap was expected")
}

func Test_calculateEndDate(t *testing.T) {
	type args struct {
		startYear     string
		lengthOfAward int
	}
	tests := []struct {
		name string
		args args
		want pgtype.Date
	}{
		{
			name: "returns the correct end date for a three year length award",
			args: args{
				startYear:     "2024",
				lengthOfAward: 3,
			},
			want: pgtype.Date{Time: time.Date(2027, time.March, 31, 0, 0, 0, 0, time.UTC), Valid: true},
		},
		{
			name: "returns the correct end date for a two year length award",
			args: args{
				startYear:     "2024",
				lengthOfAward: 2,
			},
			want: pgtype.Date{Time: time.Date(2026, time.March, 31, 0, 0, 0, 0, time.UTC), Valid: true},
		},
		{
			name: "returns the correct end date for a one year length award",
			args: args{
				startYear:     "2024",
				lengthOfAward: 1,
			},
			want: pgtype.Date{Time: time.Date(2025, time.March, 31, 0, 0, 0, 0, time.UTC), Valid: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, calculateFeeReductionEndDate(tt.args.startYear, tt.args.lengthOfAward), "calculateFeeReductionEndDate(%v, %v)", tt.args.startYear, tt.args.lengthOfAward)
		})
	}
}

func Test_calculateStartDate(t *testing.T) {
	type args struct {
		startYear string
	}
	tests := []struct {
		name string
		args args
		want pgtype.Date
	}{
		{
			name: "returns the correct start date",
			args: args{
				startYear: "2024",
			},
			want: pgtype.Date{Time: time.Date(2024, time.April, 01, 0, 0, 0, 0, time.UTC), Valid: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, calculateFeeReductionStartDate(tt.args.startYear), "calculateFeeReductionStartDate(%v)", tt.args.startYear)
		})
	}
}
