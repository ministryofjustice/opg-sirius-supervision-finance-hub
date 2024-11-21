package service

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/testhelpers"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func addFeeReductionSetup(conn testhelpers.TestConn) (Service, shared.AddFeeReduction) {
	receivedDate := shared.NewDate("2024-01-01")

	params := shared.AddFeeReduction{
		FeeType:       shared.FeeReductionTypeRemission,
		StartYear:     "2021",
		LengthOfAward: 3,
		DateReceived:  &receivedDate,
		Notes:         "Testing",
	}

	client := SetUpTest()
	s := NewService(client, conn.Conn, nil, nil)

	return s, params
}

func (suite *IntegrationSuite) TestService_AddFeeReduction() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (22, 22, '1234', 'DEMANDED', NULL);",
		"INSERT INTO fee_reduction VALUES (22, 22, 'REMISSION', NULL, '2019-04-01', '2021-03-31', 'Remission to see the notes', FALSE, '2019-05-01');",
	)
	ctx := suite.ctx

	s, params := addFeeReductionSetup(conn)
	err := s.AddFeeReduction(ctx, 22, params)

	row := conn.QueryRow(ctx, "SELECT id, finance_client_id, type, startdate, enddate, notes, datereceived, created_by, created_at FROM supervision_finance.fee_reduction WHERE id = 1")

	var (
		id            int
		financeClient int
		feeType       string
		startDate     time.Time
		endDate       time.Time
		notes         string
		dateReceived  time.Time
		createdBy     int
		createdDate   time.Time
	)

	_ = row.Scan(&id, &financeClient, &feeType, &startDate, &endDate, &notes, &dateReceived, &createdBy, &createdDate)

	assert.Equal(suite.T(), 1, id)
	assert.Equal(suite.T(), 22, financeClient)
	assert.Equal(suite.T(), "REMISSION", feeType)
	assert.Equal(suite.T(), "2021-04-01", startDate.Format("2006-01-02"))
	assert.Equal(suite.T(), "2024-03-31", endDate.Format("2006-01-02"))
	assert.Equal(suite.T(), params.Notes, notes)
	assert.Equal(suite.T(), "2024-01-01", dateReceived.Format("2006-01-02"))
	assert.Equal(suite.T(), 1, createdBy)
	assert.NotEqual(suite.T(), createdDate, "0001-01-01")

	if err == nil {
		return
	}
	suite.T().Error("Add fee reduction failed")
}

func (suite *IntegrationSuite) TestService_AddFeeReductionOverlap() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (23, 23, '1234', 'DEMANDED', NULL);",
		"INSERT INTO fee_reduction VALUES (23, 23, 'REMISSION', NULL, '2019-04-01', '2021-03-31', 'Remission to see the notes', FALSE, '2019-05-01');",
	)
	s, params := addFeeReductionSetup(conn)

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
		suite.T().Run(tc.testName, func(t *testing.T) {
			params.StartYear = tc.startYear
			params.LengthOfAward = tc.lengthOfAward

			err := s.AddFeeReduction(suite.ctx, 23, params)
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

func Test_calculateFeeReduction(t *testing.T) {
	type args struct {
		feeReductionType      shared.FeeReductionType
		invoiceTotal          int32
		invoiceFeeType        string
		generalSupervisionFee int32
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		{
			name: "remission - AD",
			args: args{
				feeReductionType:      shared.FeeReductionTypeRemission,
				invoiceTotal:          10000,
				invoiceFeeType:        "AD",
				generalSupervisionFee: 0,
			},
			want: 5000,
		},
		{
			name: "remission - non-AD - general supervision fee",
			args: args{
				feeReductionType:      shared.FeeReductionTypeRemission,
				invoiceTotal:          10000,
				invoiceFeeType:        "S2",
				generalSupervisionFee: 5000,
			},
			want: 2500,
		},
		{
			name: "remission - non-AD - no general supervision fee",
			args: args{
				feeReductionType:      shared.FeeReductionTypeRemission,
				invoiceTotal:          10000,
				invoiceFeeType:        "S2",
				generalSupervisionFee: 0,
			},
			want: 0,
		},
		{
			name: "hardship",
			args: args{
				feeReductionType:      shared.FeeReductionTypeHardship,
				invoiceTotal:          10000,
				invoiceFeeType:        "AD",
				generalSupervisionFee: 0,
			},
			want: 10000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, calculateFeeReduction(tt.args.feeReductionType, tt.args.invoiceTotal, tt.args.invoiceFeeType, tt.args.generalSupervisionFee), "calculateFeeReduction(%v, %v, %v, %v)", tt.args.feeReductionType, tt.args.invoiceTotal, tt.args.invoiceFeeType, tt.args.generalSupervisionFee)
		})
	}
}
