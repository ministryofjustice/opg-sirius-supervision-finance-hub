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
		"INSERT INTO invoice VALUES (22, 22, 22, 'S2', 'S200123/24', '2024-01-01', '2025-03-31', 10000, null, '2024-01-01', null, '2024-01-01')",
		"INSERT INTO invoice_fee_range VALUES (22, 22, 'GENERAL', '2022-04-01', '2025-03-31', 10000)",
	)
	ctx := suite.ctx

	s, params := addFeeReductionSetup(conn)
	err := s.AddFeeReduction(ctx, 22, params)

	feeReductionRow := conn.QueryRow(ctx, "SELECT id, finance_client_id, type, startdate, enddate, notes, datereceived, created_by, created_at FROM supervision_finance.fee_reduction WHERE id = 1")
	remissionLedgerRow := conn.QueryRow(ctx, "SELECT l.amount, l.notes, l.type FROM ledger_allocation la LEFT JOIN ledger l ON l.id = la.ledger_id WHERE invoice_id = 22")

	var remissionLedger struct {
		amount     int
		notes      string
		ledgerType string
	}

	_ = remissionLedgerRow.Scan(&remissionLedger.amount, &remissionLedger.notes, &remissionLedger.ledgerType)

	var feeReduction struct {
		id            int
		financeClient int
		feeType       string
		startDate     time.Time
		endDate       time.Time
		notes         string
		dateReceived  time.Time
		createdBy     int
		createdDate   time.Time
	}

	_ = feeReductionRow.Scan(
		&feeReduction.id,
		&feeReduction.financeClient,
		&feeReduction.feeType,
		&feeReduction.startDate,
		&feeReduction.endDate,
		&feeReduction.notes,
		&feeReduction.dateReceived,
		&feeReduction.createdBy,
		&feeReduction.createdDate,
	)

	assert.Equal(suite.T(), 1, feeReduction.id)
	assert.Equal(suite.T(), 22, feeReduction.financeClient)
	assert.Equal(suite.T(), "REMISSION", feeReduction.feeType)
	assert.Equal(suite.T(), "2021-04-01", feeReduction.startDate.Format("2006-01-02"))
	assert.Equal(suite.T(), "2024-03-31", feeReduction.endDate.Format("2006-01-02"))
	assert.Equal(suite.T(), params.Notes, feeReduction.notes)
	assert.Equal(suite.T(), "2024-01-01", feeReduction.dateReceived.Format("2006-01-02"))
	assert.Equal(suite.T(), 1, feeReduction.createdBy)
	assert.NotEqual(suite.T(), feeReduction.createdDate, "0001-01-01")

	assert.Equal(suite.T(), 5000, remissionLedger.amount)
	assert.Equal(suite.T(), "Credit due to approved remission", remissionLedger.notes)
	assert.Equal(suite.T(), "CREDIT REMISSION", remissionLedger.ledgerType)

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
			name: "remission - GA",
			args: args{
				feeReductionType:      shared.FeeReductionTypeRemission,
				invoiceTotal:          10000,
				invoiceFeeType:        "GA",
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
