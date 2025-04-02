package db

import (
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/testhelpers"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func (suite *IntegrationSuite) Test_invoice_adjustments() {
	ctx := suite.ctx
	today := suite.seeder.Today()
	fourYearsAgo := suite.seeder.Today().Sub(4, 0, 0)

	// one client with two orders, one with a credit memo:
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12345678", "1234")
	suite.seeder.CreateOrder(ctx, client1ID, "ACTIVE")
	_, _ = suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeGA, nil, today.StringPtr(), nil, nil, nil, nil)
	invoiceId, client1Invoice2Ref := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, today.StringPtr(), nil, nil, nil, nil)
	suite.seeder.AddFeeRanges(ctx, invoiceId, []testhelpers.FeeRange{{FromDate: today.Date(), ToDate: today.Date(), SupervisionLevel: "AD", Amount: 0}})
	suite.seeder.CreateAdjustment(ctx, client1ID, invoiceId, shared.AdjustmentTypeCreditMemo, 10000, "£100 credit", nil)

	// one client with two orders and a remission:
	client2ID := suite.seeder.CreateClient(ctx, "Barry", "Giggle", "87654321", "4321")
	suite.seeder.CreateOrder(ctx, client1ID, "ACTIVE")
	invoiceId, client2Invoice2Ref := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeGA, nil, today.StringPtr(), nil, nil, nil, nil)
	suite.seeder.AddFeeRanges(ctx, invoiceId, []testhelpers.FeeRange{{FromDate: today.Date(), ToDate: today.Date(), SupervisionLevel: "MINIMAL", Amount: 0}})
	invoiceId, _ = suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, today.StringPtr(), nil, nil, nil, nil)
	suite.seeder.AddFeeRanges(ctx, invoiceId, []testhelpers.FeeRange{{FromDate: today.Date(), ToDate: today.Date(), SupervisionLevel: "AD", Amount: 0}})
	suite.seeder.CreateFeeReduction(ctx, client2ID, shared.FeeReductionTypeRemission, strconv.Itoa(today.Date().Year()-1), 4, "Test remission", time.Now())

	c := Client{suite.seeder.Conn}

	from := shared.NewDate(fourYearsAgo.String())
	to := shared.NewDate(today.Add(0, 0, 1).String())

	rows, err := c.Run(ctx, &InvoiceAdjustments{FromDate: &from, ToDate: &to})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	formattedFinancialYear := ""

	if len(today.FinancialYear()) >= 7 {
		formattedFinancialYear = today.FinancialYear()[2:]
	}

	// client 1
	assert.Equal(suite.T(), "Ian Test", results[0]["Customer Name"], "Customer Name - client 1")
	assert.Equal(suite.T(), "12345678", results[0]["Customer number"], "Customer number - client 1")
	assert.Equal(suite.T(), "1234", results[0]["SOP number"], "SOP number - client 1")
	assert.Equal(suite.T(), "=\"0470\"", results[0]["Entity"], "Entity - client 1")
	assert.Equal(suite.T(), "10482009", results[0]["Revenue cost centre"], "Cost centre - client 1")
	assert.Equal(suite.T(), "Supervision Investigations", results[0]["Revenue cost centre description"], "Cost centre description - client 1")
	assert.Equal(suite.T(), "4481102093", results[0]["Revenue account code"], "Account code - client 1")
	assert.Equal(suite.T(), "INC - RECEIPT OF FEES AND CHARGES - Appoint Deputy", results[0]["Revenue account descriptions"], "Account code description - client 1")
	assert.Equal(suite.T(), fmt.Sprintf("MCR%s", client1Invoice2Ref), results[0]["Txn number and type"], "Txn number - client 1")
	assert.Equal(suite.T(), "Manual Credit", results[0]["Txn description"], "Txn description - client 1")
	assert.Equal(suite.T(), "", results[0]["Remission/exemption term"], "Remission/Exemption award term - client 1")
	assert.Equal(suite.T(), formattedFinancialYear, results[0]["Financial Year"], "Financial Year - client 1")
	assert.Equal(suite.T(), time.Now().Format("2006-01-02"), results[0]["Approved date"], "Approved date - client 1")
	assert.Equal(suite.T(), "100.00", results[0]["Adjustment amount"], "Adjustment amount - client 1")
	assert.Equal(suite.T(), "£100 credit", results[0]["Reason for adjustment"], "Reason for adjustment - client 1")

	// client 2
	assert.Equal(suite.T(), "Barry Giggle", results[1]["Customer Name"], "Customer Name - client 2")
	assert.Equal(suite.T(), "87654321", results[1]["Customer number"], "Customer number - client 2")
	assert.Equal(suite.T(), "4321", results[1]["SOP number"], "SOP number - client 2")
	assert.Equal(suite.T(), "=\"0470\"", results[1]["Entity"], "Entity - client 2")
	assert.Equal(suite.T(), "10482009", results[1]["Revenue cost centre"], "Cost centre - client 2")
	assert.Equal(suite.T(), "Supervision Investigations", results[1]["Revenue cost centre description"], "Cost centre description - client 2")
	assert.Equal(suite.T(), "4481102120", results[1]["Revenue account code"], "Account code - client 2")
	assert.Equal(suite.T(), "INC - RECEIPT OF FEES AND CHARGES - Rem Annual Admin Fee 3", results[1]["Revenue account descriptions"], "Account code description - client 2")
	assert.Equal(suite.T(), fmt.Sprintf("ZR%s", client2Invoice2Ref), results[1]["Txn number and type"], "Txn number - client 2")
	assert.Equal(suite.T(), "Remission Credit", results[1]["Txn description"], "Txn description - client 2")
	assert.Equal(suite.T(), "3 year", results[1]["Remission/exemption term"], "Remission/Exemption award term - client 2")
	assert.Equal(suite.T(), formattedFinancialYear, results[1]["Financial Year"], "Financial Year - client 2")
	assert.Equal(suite.T(), time.Now().Format("2006-01-02"), results[1]["Approved date"], "Approved date - client 2")
	assert.Equal(suite.T(), "100.00", results[1]["Adjustment amount"], "Adjustment amount - client 2")
	assert.Equal(suite.T(), "Test remission", results[1]["Reason for adjustment"], "Reason for adjustment - client 2")
}

func Test_invoiceAdjustments_getParams(t *testing.T) {
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
			invoiceAdjustments := &InvoiceAdjustments{
				FromDate:   tt.fromDate,
				ToDate:     tt.toDate,
				GoLiveDate: goLiveDate,
			}
			params := invoiceAdjustments.GetParams()
			assert.Equal(t, tt.expected, params)
		})
	}
}
