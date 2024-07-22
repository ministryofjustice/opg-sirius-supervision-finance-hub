package service

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
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
		InvoiceType:      shared.InvoiceTypeS2,
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
		"INSERT INTO finance_client VALUES (24, 24, '1234', 'DEMANDED', NULL);",
		"INSERT INTO fee_reduction VALUES (24, 24, 'REMISSION', NULL, '2023-04-01', '2024-03-31', 'Remission to see the notes', FALSE, '2023-05-01');",
	)
	ctx := suite.ctx
	s, params := addManualInvoiceSetup(conn)

	err := s.AddManualInvoice(ctx, 24, params)
	rows := conn.QueryRow(ctx, "SELECT * FROM supervision_finance.invoice WHERE id = 1")

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
		&createdById)

	assert.Equal(suite.T(), shared.InvoiceTypeS2.Key(), feeType)
	assert.Equal(suite.T(), 50000, amount)
	assert.Equal(suite.T(), "2024-04-12", startDate.Format("2006-01-02"))
	assert.Equal(suite.T(), "2025-03-31", endDate.Format("2006-01-02"))

	if err == nil {
		return
	}
	suite.T().Error("Add manual invoice failed")
}

func (suite *IntegrationSuite) TestService_AddManualInvoiceRaisedDateForAnInvoiceReturnsErrorForInvalidDates() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (24, 24, '1234', 'DEMANDED', NULL);",
		"INSERT INTO fee_reduction VALUES (24, 24, 'REMISSION', NULL, '2023-04-01', '2024-03-31', 'Remission to see the notes', FALSE, '2023-05-01');",
	)
	s, params := addManualInvoiceSetup(conn)

	params.RaisedDate = &shared.Date{Time: time.Now().AddDate(0, 0, 1)}
	params.StartDate = &shared.Date{Time: time.Now().AddDate(0, 0, 1)}
	params.EndDate = &shared.Date{Time: time.Now().AddDate(0, 0, -1)}
	params.InvoiceType = shared.InvoiceTypeSO

	err := s.AddManualInvoice(suite.ctx, 24, params)
	if err != nil {
		assert.Equalf(suite.T(), "bad requests: RaisedDateForAnInvoice, StartDate, EndDate", err.Error(), "Raised date %s is not in the past", params.RaisedDate)
		return
	}
}

func (suite *IntegrationSuite) TestService_AddManualInvoiceRaisedDateForAnInvoiceReturnsNoError() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (24, 24, '1234', 'DEMANDED', NULL);",
		"INSERT INTO fee_reduction VALUES (24, 24, 'REMISSION', NULL, '2023-04-01', '2024-03-31', 'Remission to see the notes', FALSE, '2023-05-01');",
	)
	s, params := addManualInvoiceSetup(conn)
	params.RaisedDate = &shared.Date{Time: time.Now().AddDate(0, 0, -1)}
	params.InvoiceType = shared.InvoiceTypeSO

	err := s.AddManualInvoice(suite.ctx, 24, params)
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

func (suite *IntegrationSuite) TestService_AddLedgerAndAllocationsForAnADInvoice() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (25, 25, '1234', 'DEMANDED', NULL);",
		"INSERT INTO fee_reduction VALUES (25, 25, 'REMISSION', NULL, '2023-04-01', '2024-03-31', 'Remission to see the notes', FALSE, '2023-05-01');",
	)

	ctx := suite.ctx
	s, params := addManualInvoiceSetup(conn)

	params.InvoiceType = shared.InvoiceTypeAD
	dateString := "2023-05-01"
	date, _ := time.Parse("2006-01-02", dateString)
	dateReceivedTransformed := &shared.Date{Time: date}

	params.StartDate = dateReceivedTransformed
	params.EndDate = dateReceivedTransformed
	params.RaisedDate = dateReceivedTransformed

	err := s.AddManualInvoice(ctx, 25, params)
	if err != nil {
		suite.T().Error("Add manual invoice ledger failed")
	}
	var ledger store.Ledger
	q := conn.QueryRow(ctx, "SELECT id, amount, notes, type, status, finance_client_id FROM ledger WHERE id = 1")
	err = q.Scan(&ledger.ID, &ledger.Amount, &ledger.Notes, &ledger.Type, &ledger.Status, &ledger.FinanceClientID)
	if err != nil {
		suite.T().Error("Add manual invoice ledger failed")
	} else {
		expected := store.Ledger{
			ID:              1,
			Amount:          int32(params.Amount / 2),
			Notes:           pgtype.Text{String: "Credit due to manual invoice REMISSION", Valid: true},
			Type:            "CREDIT REMISSION",
			Status:          "APPROVED",
			FinanceClientID: pgtype.Int4{Int32: int32(25), Valid: true},
		}

		assert.EqualValues(suite.T(), expected, ledger)
	}
}

func (suite *IntegrationSuite) TestService_AddLedgerAndAllocationsForAnExemption() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (25, 25, '1234', 'DEMANDED', NULL);",
		"INSERT INTO fee_reduction VALUES (25, 25, 'EXEMPTION', NULL, '2022-04-01', '2025-03-31', 'Exemption to see the notes', FALSE, '2023-05-01');",
	)

	ctx := suite.ctx
	s, params := addManualInvoiceSetup(conn)

	err := s.AddManualInvoice(ctx, 25, params)
	if err != nil {
		suite.T().Error("Add manual invoice ledger with an exemption failed")
	}
	var ledger store.Ledger
	q := conn.QueryRow(ctx, "SELECT id, amount, notes, type, status, finance_client_id FROM ledger WHERE id = 1")
	err = q.Scan(&ledger.ID, &ledger.Amount, &ledger.Notes, &ledger.Type, &ledger.Status, &ledger.FinanceClientID)
	if err != nil {
		suite.T().Error("Add manual invoice ledger with an exemption failed")
	} else {
		expected := store.Ledger{
			ID:              1,
			Amount:          int32(params.Amount),
			Notes:           pgtype.Text{String: "Credit due to manual invoice EXEMPTION", Valid: true},
			Type:            "CREDIT EXEMPTION",
			Status:          "APPROVED",
			FinanceClientID: pgtype.Int4{Int32: int32(25), Valid: true},
		}

		assert.EqualValues(suite.T(), expected, ledger)
	}
}
