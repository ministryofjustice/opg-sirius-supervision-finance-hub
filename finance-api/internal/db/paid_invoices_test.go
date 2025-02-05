package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func (suite *IntegrationSuite) Test_paid_invoices() {
	ctx := suite.ctx

	today := suite.seeder.Today()
	twoMonthsAgo := suite.seeder.Today().Sub(0, 2, 0)
	twoYearsAgo := suite.seeder.Today().Sub(2, 0, 0)
	fourYearsAgo := suite.seeder.Today().Sub(4, 0, 0)
	general := "320.00"

	// one client with one invoice and an exemption
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12345678", "1234")
	suite.seeder.CreateOrder(ctx, client1ID, "ACTIVE")
	_, c1i1Ref := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreateFeeReduction(ctx, client1ID, shared.FeeReductionTypeExemption, strconv.Itoa(twoYearsAgo.Date().Year()), 2, "Test exemption", time.Now())

	// one client with:
	// one invoice with no outstanding balance due to an exemption
	// one invoice with outstanding balance
	client2ID := suite.seeder.CreateClient(ctx, "John", "Suite", "87654321", "4321")
	suite.seeder.CreateOrder(ctx, client2ID, "ACTIVE")
	_, c2i1Ref := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeAD, nil, fourYearsAgo.StringPtr(), nil, nil, nil, nil)
	_, _ = suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS2, &general, twoMonthsAgo.StringPtr(), twoMonthsAgo.StringPtr(), nil, nil, nil)
	suite.seeder.CreateFeeReduction(ctx, client2ID, shared.FeeReductionTypeExemption, strconv.Itoa(fourYearsAgo.Date().Year()-1), 2, "Test exemption", time.Now())

	// one client with one invoice partially paid due to a remission
	client3ID := suite.seeder.CreateClient(ctx, "John", "Suite", "87654321", "4321")
	suite.seeder.CreateOrder(ctx, client3ID, "ACTIVE")
	_, _ = suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeAD, nil, fourYearsAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreateFeeReduction(ctx, client3ID, shared.FeeReductionTypeRemission, strconv.Itoa(fourYearsAgo.Date().Year()-1), 4, "Test remission", time.Now())

	c := Client{suite.seeder.Conn}

	from := shared.NewDate(fourYearsAgo.String())
	to := shared.NewDate(today.String())

	rows, err := c.Run(ctx, &PaidInvoices{
		FromDate: &from,
		ToDate:   &to,
	})

	runTime := time.Now()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// client 2 invoice 1
	assert.Equal(suite.T(), "John Suite", results[0]["Customer name"], "Customer Name - client 2 invoice 1")
	assert.Equal(suite.T(), "87654321", results[0]["Customer number"], "Customer number - client 2 invoice 1")
	assert.Equal(suite.T(), "4321", results[0]["SOP number"], "SOP number - client 2 invoice 1")
	assert.Equal(suite.T(), "=\"0470\"", results[0]["Entity"], "Entity - client 2 invoice 1")
	assert.Equal(suite.T(), "10482009", results[0]["Cost centre"], "Cost centre - client 2 invoice 1")
	assert.Equal(suite.T(), "Supervision Investigations", results[0]["Cost centre description"], "Cost centre description - client 2 invoice 1")
	assert.Equal(suite.T(), "4481102114", results[0]["Account code"], "Account code - client 2 invoice 1")
	assert.Equal(suite.T(), "INC - RECEIPT OF FEES AND CHARGES - Rem Appoint Deputy", results[0]["Account code description"], "Account code description - client 2 invoice 1")
	assert.Equal(suite.T(), "AD", results[0]["Invoice type"], "Invoice type - client 2 invoice 1")
	assert.Equal(suite.T(), c2i1Ref, results[0]["Invoice number"], "Invoice number - client 2 invoice 1")
	assert.Equal(suite.T(), "ZE"+c2i1Ref, results[0]["Txn number"], "Txn number - client 2 invoice 1")
	assert.Equal(suite.T(), "Exemption Credit", results[0]["Txn description"], "Txn description - client 2 invoice 1")
	assert.Equal(suite.T(), "100.00", results[0]["Original amount"], "Original amount - client 2 invoice 1")
	assert.Equal(suite.T(), "<nil>", results[0]["Received date"], "Received date - client 2 invoice 1")
	assert.Contains(suite.T(), results[0]["Sirius upload date"], runTime.Format("2006-01-02 15:04"), "Sirius upload date - client 2 invoice 1")
	assert.Equal(suite.T(), "0", results[0]["Cash amount"], "Cash amount - client 2 invoice 1")
	assert.Equal(suite.T(), "100.00", results[0]["Credit amount"], "Credit amount - client 2 invoice 1")
	assert.Equal(suite.T(), "0", results[0]["Adjustment amount"], "Adjustment amount - client 2 invoice 1")
	assert.Equal(suite.T(), "Test exemption", results[0]["Memo line description"], "Memo line description - client 2 invoice 1")

	// client 1
	assert.Equal(suite.T(), "Ian Test", results[1]["Customer name"], "Customer Name - client 1")
	assert.Equal(suite.T(), "12345678", results[1]["Customer number"], "Customer number - client 1")
	assert.Equal(suite.T(), "1234", results[1]["SOP number"], "SOP number - client 1")
	assert.Equal(suite.T(), "=\"0470\"", results[1]["Entity"], "Entity - client 1")
	assert.Equal(suite.T(), "10482009", results[1]["Cost centre"], "Cost centre - client 1")
	assert.Equal(suite.T(), "Supervision Investigations", results[1]["Cost centre description"], "Cost centre description - client 1")
	assert.Equal(suite.T(), "4481102114", results[1]["Account code"], "Account code - client 1")
	assert.Equal(suite.T(), "INC - RECEIPT OF FEES AND CHARGES - Rem Appoint Deputy", results[1]["Account code description"], "Account code description - client 1")
	assert.Equal(suite.T(), "AD", results[1]["Invoice type"], "Invoice type - client 1")
	assert.Equal(suite.T(), c1i1Ref, results[1]["Invoice number"], "Invoice number - client 1")
	assert.Equal(suite.T(), "ZE"+c1i1Ref, results[1]["Txn number"], "Txn number - client 1")
	assert.Equal(suite.T(), "Exemption Credit", results[1]["Txn description"], "Txn description - client 1")
	assert.Equal(suite.T(), "100.00", results[1]["Original amount"], "Original amount - client 1")
	assert.Equal(suite.T(), "<nil>", results[1]["Received date"], "Received date - client 1")
	assert.Contains(suite.T(), results[1]["Sirius upload date"], runTime.Format("2006-01-02 15:04"), "Sirius upload date - client 1")
	assert.Equal(suite.T(), "0", results[1]["Cash amount"], "Cash amount - client 1")
	assert.Equal(suite.T(), "100.00", results[1]["Credit amount"], "Credit amount - client 1")
	assert.Equal(suite.T(), "0", results[1]["Adjustment amount"], "Adjustment amount - client 1")
	assert.Equal(suite.T(), "Test exemption", results[1]["Memo line description"], "Memo line description - client 1")
}

func Test_paidInvoices_getParams(t *testing.T) {
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
			paidInvoices := &PaidInvoices{
				FromDate:   tt.fromDate,
				ToDate:     tt.toDate,
				GoLiveDate: goLiveDate,
			}
			params := paidInvoices.GetParams()
			assert.Equal(t, tt.expected, params)
		})
	}
}
