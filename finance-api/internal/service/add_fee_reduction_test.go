package service

import (
	"context"
	"database/sql"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/finance-api/internal/testhelpers"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func setupServiceAndParams(conn testhelpers.TestConn) (*Service, shared.AddFeeReduction) {
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
		Store: store.New(conn),
		TX:    conn,
	}

	return s, params
}

func TestService_AddFeeReduction(t *testing.T) {
	conn := testDB.GetConn()
	t.Cleanup(func() {
		testDB.Restore()
	})

	conn.SeedData(
		"INSERT INTO finance_client VALUES (22, 22, '1234', 'DEMANDED', null, 12300, 2222);",
		"INSERT INTO fee_reduction VALUES (22, 22, 'REMISSION', null, '2019-04-01', '2021-03-31', 'Remission to see the notes', false, '2019-05-01');",
	)
	ctx := context.Background()
	s, params := setupServiceAndParams(conn)

	err := s.AddFeeReduction(22, params)
	rows, _ := conn.Query(ctx, "SELECT * FROM supervision_finance.fee_reduction WHERE id = 1")
	defer rows.Close()

	for rows.Next() {
		var (
			id            int
			financeClient int
			feeType       string
			evidenceType  sql.NullString
			startDate     time.Time
			endDate       time.Time
			notes         string
			deleted       bool
			dateReceived  time.Time
		)

		_ = rows.Scan(&id, &financeClient, &feeType, &evidenceType, &startDate, &endDate, &notes, &deleted, &dateReceived)

		assert.Equal(t, "REMISSION", feeType)
		assert.Equal(t, "2021-04-01", startDate.Format("2006-01-02"))
		assert.Equal(t, "2024-03-31", endDate.Format("2006-01-02"))
		assert.Equal(t, params.Notes, notes)
	}

	if err == nil {
		return
	}
	t.Error("Add fee reduction failed")
}

func TestService_AddFeeReductionOverlap(t *testing.T) {
	conn := testDB.GetConn()
	t.Cleanup(func() {
		testDB.Restore()
	})

	conn.SeedData(
		"INSERT INTO finance_client VALUES (23, 23, '1234', 'DEMANDED', null, 12300, 2222);",
		"INSERT INTO fee_reduction VALUES (23, 23, 'REMISSION', null, '2019-04-01', '2021-03-31', 'Remission to see the notes', false, '2019-05-01');",
	)
	s, params := setupServiceAndParams(conn)

	testCases := []struct {
		testName      string
		startYear     string
		lengthOfAward int
	}{
		{
			testName:      "Overlap starts one year before existing",
			startYear:     "2018",
			lengthOfAward: 2,
		},
		{
			testName:      "Overlap starts the same date as existing",
			startYear:     "2019",
			lengthOfAward: 1,
		},
		{
			testName:      "Overlap end date is the same as existing",
			startYear:     "2018",
			lengthOfAward: 3,
		},
		{
			testName:      "Overlap both dates are different to existing and overlap",
			startYear:     "2020",
			lengthOfAward: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			params.StartYear = tc.startYear
			params.LengthOfAward = tc.lengthOfAward

			err := s.AddFeeReduction(23, params)
			if err != nil {
				assert.Equalf(t, "overlap", err.Error(), "StartYear %s has an overlap", tc.startYear)
				return
			}
			t.Error("Overlap was expected")
		})
	}
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
