package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func (suite *IntegrationSuite) Test_bad_debt_write_off() {
	ctx := suite.ctx
	today := suite.seeder.Today()
	twoMonthsAgo := today.Sub(0, 2, 0)
	twoYearsAgo := today.Sub(2, 0, 0)
	fourYearsAgo := today.Sub(4, 0, 0)
	general := "320.00"

	suite.seeder.CreateTestAssignee(ctx)

	// one client with:
	// - one written off invoice
	// - one active invoice
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12345678", "1234")
	suite.seeder.CreateOrder(ctx, client1ID, "ACTIVE")
	_, _ = suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeGA, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, nil)
	paidInvoiceID, c1i1Ref := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreateAdjustment(ctx, client1ID, paidInvoiceID, shared.AdjustmentTypeWriteOff, 0, "Written off", nil)

	// one client with two written off invoices
	client2ID := suite.seeder.CreateClient(ctx, "John", "Suite", "87654321", "4321")
	suite.seeder.CreateOrder(ctx, client2ID, "ACTIVE")
	paidInvoiceID, c2i1Ref := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeAD, nil, fourYearsAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreateAdjustment(ctx, client1ID, paidInvoiceID, shared.AdjustmentTypeWriteOff, 0, "Written off", nil)

	paidInvoiceID, c2i2Ref := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS2, &general, twoYearsAgo.StringPtr(), twoYearsAgo.StringPtr(), nil, nil, nil)
	suite.seeder.CreateAdjustment(ctx, client1ID, paidInvoiceID, shared.AdjustmentTypeWriteOff, 0, "Written off", nil)

	// one client with one unapproved write off
	client3ID := suite.seeder.CreateClient(ctx, "John", "Suite", "87654321", "4321")
	suite.seeder.CreateOrder(ctx, client3ID, "ACTIVE")
	paidInvoiceID, _ = suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeAD, nil, fourYearsAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePendingAdjustment(ctx, client1ID, paidInvoiceID, shared.AdjustmentTypeWriteOff, 0, "Written off")

	c := Client{suite.seeder.Conn}

	from := shared.NewDate(fourYearsAgo.String())
	to := shared.NewDate(today.String())

	rows, err := c.Run(ctx, NewBadDebtWriteOff(&from, &to, time.Time{}))

	runTime := today

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 4, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// client 1
	assert.Equal(suite.T(), "Ian Test", results[0]["Customer name"], "Customer Name - client 1")
	assert.Equal(suite.T(), "12345678", results[0]["Customer number"], "Customer number - client 1")
	assert.Equal(suite.T(), "1234", results[0]["SOP number"], "SOP number - client 1")
	assert.Equal(suite.T(), "=\"0470\"", results[0]["Entity"], "Entity - client 1")
	assert.Equal(suite.T(), "10482009", results[0]["Cost centre"], "Cost centre - client 1")
	assert.Equal(suite.T(), "5356202100", results[0]["Account code"], "Account code - client 1")
	assert.Equal(suite.T(), "EXP - IMPAIRMENT - BAD DEBTS-Appoint Deputy Write Off", results[0]["Account code description"], "Account code description - client 1")
	assert.Equal(suite.T(), "100.00", results[0]["Adjustment amount"], "Adjustment amount - client 1")
	assert.Contains(suite.T(), results[0]["Adjustment date"], runTime.Date().Format("2006-01-02 15:04"), "Adjustment date - client 1")
	assert.Equal(suite.T(), "WO"+c1i1Ref, results[0]["Txn number"], "Txn number - client 1")
	assert.Equal(suite.T(), "Johnny Test", results[0]["Approver"], "Approver - client 1")

	// client 2 - write off 1
	assert.Equal(suite.T(), "Ian Test", results[1]["Customer name"], "Customer Name - client 2 write off 1")
	assert.Equal(suite.T(), "12345678", results[1]["Customer number"], "Customer number - client 2 write off 1")
	assert.Equal(suite.T(), "1234", results[1]["SOP number"], "SOP number - client 2 write off 1")
	assert.Equal(suite.T(), "=\"0470\"", results[1]["Entity"], "Entity - client 2 write off 1")
	assert.Equal(suite.T(), "10482009", results[1]["Cost centre"], "Cost centre - client 2 write off 1")
	assert.Equal(suite.T(), "5356202100", results[1]["Account code"], "Account code - client 2 write off 1")
	assert.Equal(suite.T(), "EXP - IMPAIRMENT - BAD DEBTS-Appoint Deputy Write Off", results[1]["Account code description"], "Account code description - client 2 write off 1")
	assert.Equal(suite.T(), "100.00", results[1]["Adjustment amount"], "Adjustment amount - client 2 write off 1")
	assert.Contains(suite.T(), results[1]["Adjustment date"], runTime.Date().Format("2006-01-02 15:04"), "Adjustment date - client 2 write off 1")
	assert.Equal(suite.T(), "WO"+c2i1Ref, results[1]["Txn number"], "Txn number - client 2 write off 1")
	assert.Equal(suite.T(), "Johnny Test", results[1]["Approver"], "Approver - client 2 write off 1")

	// client 2 - write off 2
	assert.Equal(suite.T(), "Ian Test", results[2]["Customer name"], "Customer Name - client 2 write off 2")
	assert.Equal(suite.T(), "12345678", results[2]["Customer number"], "Customer number - client 2 write off 2")
	assert.Equal(suite.T(), "1234", results[2]["SOP number"], "SOP number - client 2 write off 2")
	assert.Equal(suite.T(), "=\"0470\"", results[2]["Entity"], "Entity - client 2 write off 2")
	assert.Equal(suite.T(), "10482009", results[2]["Cost centre"], "Cost centre - client 2 write off 2")
	assert.Equal(suite.T(), "5356202102", results[2]["Account code"], "Account code - client 2 write off 2")
	assert.Equal(suite.T(), "EXP - IMPAIRMENT - BAD DEBTS-Sup Fee 2 Write Off\tWrite-off", results[2]["Account code description"], "Account code description - client 2 write off 2")
	assert.Equal(suite.T(), "320.00", results[2]["Adjustment amount"], "Adjustment amount - client 2 write off 2")
	assert.Contains(suite.T(), results[2]["Adjustment date"], runTime.Date().Format("2006-01-02 15:04"), "Adjustment date - client 2 write off 2")
	assert.Equal(suite.T(), "WO"+c2i2Ref, results[2]["Txn number"], "Txn number - client 2 write off 2")
	assert.Equal(suite.T(), "Johnny Test", results[2]["Approver"], "Approver - client 2 write off 2")
}

func Test_badDebtWriteOff_getParams(t *testing.T) {
	today := time.Now()
	goLiveDate := today.AddDate(-4, 0, 0)
	toDate := shared.NewDate(today.AddDate(-1, 0, 0).Format("2006-01-02"))
	fromDate := shared.NewDate(today.AddDate(-2, 0, 0).Format("2006-01-02"))

	tests := []struct {
		name     string
		fromDate *shared.Date
		toDate   *shared.Date
		expected []any
	}{
		{
			name:     "No FromDate and ToDate",
			fromDate: nil,
			toDate:   nil,
			expected: []any{goLiveDate.Format("2006-01-02"), today.Format("2006-01-02")},
		},
		{
			name:     "With FromDate and ToDate",
			fromDate: &toDate,
			toDate:   &fromDate,
			expected: []any{toDate.Time.Format("2006-01-02"), fromDate.Time.Format("2006-01-02")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			badDebtWriteOff := &BadDebtWriteOff{
				FromDate:   tt.fromDate,
				ToDate:     tt.toDate,
				GoLiveDate: goLiveDate,
			}
			params := badDebtWriteOff.GetParams()
			assert.Equal(t, tt.expected, params)
		})
	}
}
