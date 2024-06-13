package service

import (
	"context"
	"github.com/opg-sirius-finance-hub/finance-api/internal/testhelpers"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func addManualInvoiceSetup(conn testhelpers.TestConn) (Service, shared.AddManualInvoice) {
	var startDateTransformed *shared.Date
	var endDateTransformed *shared.Date

	startDateToTime, _ := time.Parse("2006-01-02", "2024-04-12")
	startDateTransformed = &shared.Date{Time: startDateToTime}

	endDateToTime, _ := time.Parse("2006-01-02", "2025-03-31")
	endDateTransformed = &shared.Date{Time: endDateToTime}

	params := shared.AddManualInvoice{
		InvoiceType:      "S2",
		Amount:           50000,
		RaisedDate:       endDateTransformed,
		StartDate:        startDateTransformed,
		EndDate:          endDateTransformed,
		SupervisionLevel: "GENERAL",
	}

	s := NewService(conn.Conn)

	return s, params
}

func (suite *IntegrationSuite) TestService_AddManualInvoice() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (24, 24, '1234', 'DEMANDED', null, 12300, 2222);",
		"INSERT INTO fee_reduction VALUES (24, 24, 'REMISSION', null, '2023-04-01', '2024-03-31', 'Remission to see the notes', false, '2023-05-01');",
	)
	ctx := context.Background()
	s, params := addManualInvoiceSetup(conn)

	err := s.AddManualInvoice(24, params)
	rows, _ := conn.Query(ctx, "SELECT * FROM supervision_finance.invoice where id = 1")
	defer rows.Close()

	for rows.Next() {
		var (
			id                int
			personId          int
			financeClientId   int
			feeType           string
			reference         string
			startDate         time.Time
			endDate           time.Time
			amount            int
			supervisionLevel  string
			confirmedDate     time.Time
			batchNumber       int
			raisedDate        time.Time
			source            string
			scheduledFn14Date time.Time
			cachedDebtAmount  int
			createdDate       time.Time
			createdById       int
			feeReductionId    int
		)

		_ = rows.Scan(
			&id,
			&personId,
			&financeClientId,
			&feeType,
			&reference,
			&startDate,
			&endDate,
			&amount,
			&supervisionLevel,
			&confirmedDate,
			&batchNumber,
			&raisedDate,
			&source,
			&scheduledFn14Date,
			&cachedDebtAmount,
			&createdDate,
			&createdById,
			&feeReductionId)

		assert.Equal(suite.T(), "S2", feeType)
		assert.Equal(suite.T(), 50000, amount)
		assert.Equal(suite.T(), "2024-04-12", startDate.Format("2006-01-02"))
		assert.Equal(suite.T(), "2025-03-31", endDate.Format("2006-01-02"))
	}

	if err == nil {
		return
	}
	suite.T().Error("Add manual invoice failed")
}

func (suite *IntegrationSuite) TestService_AddManualInvoiceRaisedDateForAnInvoiceReturnsErrorForInvalidDates() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (24, 24, '1234', 'DEMANDED', null, 12300, 2222);",
		"INSERT INTO fee_reduction VALUES (24, 24, 'REMISSION', null, '2023-04-01', '2024-03-31', 'Remission to see the notes', false, '2023-05-01');",
	)
	s, params := addManualInvoiceSetup(conn)

	params.RaisedDate = &shared.Date{Time: time.Now().AddDate(0, 0, 1)}
	params.StartDate = &shared.Date{Time: time.Now().AddDate(0, 0, 1)}
	params.EndDate = &shared.Date{Time: time.Now().AddDate(0, 0, -1)}
	params.InvoiceType = "SO"

	err := s.AddManualInvoice(24, params)
	if err != nil {
		assert.Equalf(suite.T(), "bad requests: RaisedDateForAnInvoice, StartDate, EndDate", err.Error(), "Raised date %s is not in the past", params.RaisedDate)
		return

	}
}

func (suite *IntegrationSuite) TestService_AddManualInvoiceRaisedDateForAnInvoiceReturnsNoError() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (24, 24, '1234', 'DEMANDED', null, 12300, 2222);",
		"INSERT INTO fee_reduction VALUES (24, 24, 'REMISSION', null, '2023-04-01', '2024-03-31', 'Remission to see the notes', false, '2023-05-01');",
	)
	s, params := addManualInvoiceSetup(conn)
	params.RaisedDate = &shared.Date{Time: time.Now().AddDate(0, 0, -1)}
	params.InvoiceType = "SO"

	err := s.AddManualInvoice(24, params)
	if err == nil {
		return
	}
	suite.T().Error("validRaisedDateInThePast failed")
}

func TestService_AddManualInvoiceAddLeadingZeros(t *testing.T) {
	tests := []struct {
		name   string
		number int
		want   string
	}{
		{
			name:   "returns the correct padded number for one number passed in",
			number: 1,
			want:   "000001",
		},
		{
			name:   "returns the correct padded number for six number passed in",
			number: 123456,
			want:   "123456",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, addLeadingZeros(tt.number), "addLeadingZeros(%v)", tt.number)
		})
	}
}

func Test_validateEndDate(t *testing.T) {
	type args struct {
		startDate *shared.Date
		endDate   *shared.Date
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "returns true if the end date is in the future compared to start date",
			args: args{
				startDate: &shared.Date{Time: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)},
				endDate:   &shared.Date{Time: time.Date(2024, 5, 2, 0, 0, 0, 0, time.UTC)},
			},
			want: true,
		},
		{
			name: "returns false if the end date before start date",
			args: args{
				startDate: &shared.Date{Time: time.Date(2024, 5, 2, 0, 0, 0, 0, time.UTC)},
				endDate:   &shared.Date{Time: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, validateEndDate(tt.args.startDate, tt.args.endDate), "validateEndDate(%v, %v)", tt.args.startDate, tt.args.endDate)
		})
	}
}

func Test_validateStartDate(t *testing.T) {
	type args struct {
		startDate *shared.Date
		endDate   *shared.Date
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "returns true if the start date is in the past compared to end date",
			args: args{
				startDate: &shared.Date{Time: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)},
				endDate:   &shared.Date{Time: time.Date(2024, 5, 2, 0, 0, 0, 0, time.UTC)},
			},
			want: true,
		},
		{
			name: "returns false if the start date before end date",
			args: args{
				startDate: &shared.Date{Time: time.Date(2024, 5, 2, 0, 0, 0, 0, time.UTC)},
				endDate:   &shared.Date{Time: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, validateStartDate(tt.args.startDate, tt.args.endDate), "validateStartDate(%v, %v)", tt.args.startDate, tt.args.endDate)
		})
	}
}

func Test_isSameFinancialYear(t *testing.T) {
	type args struct {
		startDate *shared.Date
		endDate   *shared.Date
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "returns false if the start date and end date are not in the same financial year",
			args: args{
				startDate: &shared.Date{Time: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)},
				endDate:   &shared.Date{Time: time.Date(2025, 5, 2, 0, 0, 0, 0, time.UTC)},
			},
			want: false,
		},
		{
			name: "returns true if the start date and end date are in the same financial year",
			args: args{
				startDate: &shared.Date{Time: time.Date(2024, 4, 01, 0, 0, 0, 0, time.UTC)},
				endDate:   &shared.Date{Time: time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC)},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, isSameFinancialYear(tt.args.startDate, tt.args.endDate), "isSameFinancialYear(%v, %v)", tt.args.startDate, tt.args.endDate)
		})
	}
}
