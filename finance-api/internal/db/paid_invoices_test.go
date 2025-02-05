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
	yesterday := suite.seeder.Today().Sub(0, 0, 1)
	twoMonthsAgo := suite.seeder.Today().Sub(0, 2, 0)
	twoYearsAgo := suite.seeder.Today().Sub(2, 0, 0)
	fourYearsAgo := suite.seeder.Today().Sub(4, 0, 0)
	general := "320.00"
	minimal := "10.00"

	// client with:
	// one invoice
	// one exemption
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12345678", "1234")
	suite.seeder.CreateOrder(ctx, client1ID, "ACTIVE")
	_, c1i1Ref := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreateFeeReduction(ctx, client1ID, shared.FeeReductionTypeExemption, strconv.Itoa(twoYearsAgo.Date().Year()), 2, "Test exemption", time.Now())

	// client with:
	// one invoice with no outstanding balance due to an exemption
	// one invoice with outstanding balance
	client2ID := suite.seeder.CreateClient(ctx, "John", "Suite", "87654321", "4321")
	suite.seeder.CreateOrder(ctx, client2ID, "ACTIVE")
	_, c2i1Ref := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeAD, nil, fourYearsAgo.StringPtr(), nil, nil, nil, nil)
	_, _ = suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS2, &general, twoMonthsAgo.StringPtr(), twoMonthsAgo.StringPtr(), nil, nil, nil)
	suite.seeder.CreateFeeReduction(ctx, client2ID, shared.FeeReductionTypeExemption, strconv.Itoa(fourYearsAgo.Date().Year()-1), 2, "Test exemption", time.Now())

	// client with:
	// one invoice partially paid due to a remission
	client3ID := suite.seeder.CreateClient(ctx, "John", "Suite", "87654321", "4321")
	suite.seeder.CreateOrder(ctx, client3ID, "ACTIVE")
	_, _ = suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeAD, nil, fourYearsAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreateFeeReduction(ctx, client3ID, shared.FeeReductionTypeRemission, strconv.Itoa(fourYearsAgo.Date().Year()-1), 4, "Test remission", time.Now())

	// client with:
	//one invoice paid with supervision BACS payment
	client4ref := "11111111"
	client4ID := suite.seeder.CreateClient(ctx, "Sally", "Supervision", client4ref, "1111")
	suite.seeder.CreateOrder(ctx, client4ID, "ACTIVE")
	_, c4i1Ref := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeS3, &minimal, yesterday.StringPtr(), nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 1000, yesterday.Date(), client4ref, shared.TransactionTypeSupervisionBACSPayment, yesterday.Date())

	// client with:
	//one invoice paid with OPG BACS payment
	client5ref := "22222222"
	client5ID := suite.seeder.CreateClient(ctx, "Owen", "OPG", client5ref, "2222")
	suite.seeder.CreateOrder(ctx, client5ID, "ACTIVE")
	_, c5i1Ref := suite.seeder.CreateInvoice(ctx, client5ID, shared.InvoiceTypeS2, &general, today.StringPtr(), nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 32000, today.Date(), client5ref, shared.TransactionTypeOPGBACSPayment, today.Date())

	c := Client{suite.seeder.Conn}

	from := shared.NewDate(fourYearsAgo.String())
	to := shared.NewDate(today.String())

	rows, err := c.Run(ctx, &PaidInvoices{
		FromDate: &from,
		ToDate:   &to,
	})

	runTime := time.Now()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 5, len(rows))

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
	assert.Equal(suite.T(), "", results[0]["Received date"], "Received date - client 2 invoice 1")
	assert.Contains(suite.T(), results[0]["Sirius upload date"], runTime.Format("2006-01-02"), "Sirius upload date - client 2 invoice 1")
	assert.Equal(suite.T(), "0", results[0]["Cash amount"], "Cash amount - client 2 invoice 1")
	assert.Equal(suite.T(), "100.00", results[0]["Credit amount"], "Credit amount - client 2 invoice 1")
	assert.Equal(suite.T(), "0", results[0]["Adjustment amount"], "Adjustment amount - client 2 invoice 1")
	assert.Equal(suite.T(), "Test exemption", results[0]["Memo line description"], "Memo line description - client 2 invoice 1")

	// client 4
	assert.Equal(suite.T(), "Sally Supervision", results[1]["Customer name"], "Customer Name - client 4")
	assert.Equal(suite.T(), client4ref, results[1]["Customer number"], "Customer number - client 4")
	assert.Equal(suite.T(), "1111", results[1]["SOP number"], "SOP number - client 4")
	assert.Equal(suite.T(), "=\"0470\"", results[1]["Entity"], "Entity - client 4")
	assert.Equal(suite.T(), "99999999", results[1]["Cost centre"], "Cost centre - client 4")
	assert.Equal(suite.T(), "BALANCE SHEET", results[1]["Cost centre description"], "Cost centre description - client 4")
	assert.Equal(suite.T(), "1816100000", results[1]["Account code"], "Account code - client 4")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES", results[1]["Account code description"], "Account code description - client 4")
	assert.Equal(suite.T(), "S3", results[1]["Invoice type"], "Invoice type - client 4")
	assert.Equal(suite.T(), c4i1Ref, results[1]["Invoice number"], "Invoice number - client 4")
	assert.Equal(suite.T(), "BC"+c4i1Ref, results[1]["Txn number"], "Txn number - client 4")
	assert.Equal(suite.T(), "BACS Payment", results[1]["Txn description"], "Txn description - client 4")
	assert.Equal(suite.T(), "10.00", results[1]["Original amount"], "Original amount - client 4")
	assert.Equal(suite.T(), yesterday.String(), results[1]["Received date"], "Received date - client 4")
	assert.Contains(suite.T(), results[1]["Sirius upload date"], yesterday.String(), "Sirius upload date - client 4")
	assert.Equal(suite.T(), "10.00", results[1]["Cash amount"], "Cash amount - client 4")
	assert.Equal(suite.T(), "0", results[1]["Credit amount"], "Credit amount - client 4")
	assert.Equal(suite.T(), "0", results[1]["Adjustment amount"], "Adjustment amount - client 4")
	assert.Equal(suite.T(), "", results[1]["Memo line description"], "Memo line description - client 4")

	// client 5
	assert.Equal(suite.T(), "Owen OPG", results[2]["Customer name"], "Customer Name - client 5")
	assert.Equal(suite.T(), client5ref, results[2]["Customer number"], "Customer number - client 5")
	assert.Equal(suite.T(), "2222", results[2]["SOP number"], "SOP number - client 5")
	assert.Equal(suite.T(), "=\"0470\"", results[2]["Entity"], "Entity - client 5")
	assert.Equal(suite.T(), "99999999", results[2]["Cost centre"], "Cost centre - client 5")
	assert.Equal(suite.T(), "BALANCE SHEET", results[2]["Cost centre description"], "Cost centre description - client 5")
	assert.Equal(suite.T(), "1816100000", results[2]["Account code"], "Account code - client 5")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES", results[2]["Account code description"], "Account code description - client 5")
	assert.Equal(suite.T(), "S2", results[2]["Invoice type"], "Invoice type - client 5")
	assert.Equal(suite.T(), c5i1Ref, results[2]["Invoice number"], "Invoice number - client 5")
	assert.Equal(suite.T(), "BC"+c5i1Ref, results[2]["Txn number"], "Txn number - client 5")
	assert.Equal(suite.T(), "BACS Payment", results[2]["Txn description"], "Txn description - client 5")
	assert.Equal(suite.T(), "320.00", results[2]["Original amount"], "Original amount - client 5")
	assert.Equal(suite.T(), today.String(), results[2]["Received date"], "Received date - client 5")
	assert.Contains(suite.T(), results[2]["Sirius upload date"], today.String(), "Sirius upload date - client 5")
	assert.Equal(suite.T(), "320.00", results[2]["Cash amount"], "Cash amount - client 5")
	assert.Equal(suite.T(), "0", results[2]["Credit amount"], "Credit amount - client 5")
	assert.Equal(suite.T(), "0", results[2]["Adjustment amount"], "Adjustment amount - client 5")
	assert.Equal(suite.T(), "", results[2]["Memo line description"], "Memo line description - client 5")

	// client 1
	assert.Equal(suite.T(), "Ian Test", results[3]["Customer name"], "Customer Name - client 1")
	assert.Equal(suite.T(), "12345678", results[3]["Customer number"], "Customer number - client 1")
	assert.Equal(suite.T(), "1234", results[3]["SOP number"], "SOP number - client 1")
	assert.Equal(suite.T(), "=\"0470\"", results[3]["Entity"], "Entity - client 1")
	assert.Equal(suite.T(), "10482009", results[3]["Cost centre"], "Cost centre - client 1")
	assert.Equal(suite.T(), "Supervision Investigations", results[3]["Cost centre description"], "Cost centre description - client 1")
	assert.Equal(suite.T(), "4481102114", results[3]["Account code"], "Account code - client 1")
	assert.Equal(suite.T(), "INC - RECEIPT OF FEES AND CHARGES - Rem Appoint Deputy", results[3]["Account code description"], "Account code description - client 1")
	assert.Equal(suite.T(), "AD", results[3]["Invoice type"], "Invoice type - client 1")
	assert.Equal(suite.T(), c1i1Ref, results[3]["Invoice number"], "Invoice number - client 1")
	assert.Equal(suite.T(), "ZE"+c1i1Ref, results[3]["Txn number"], "Txn number - client 1")
	assert.Equal(suite.T(), "Exemption Credit", results[3]["Txn description"], "Txn description - client 1")
	assert.Equal(suite.T(), "100.00", results[3]["Original amount"], "Original amount - client 1")
	assert.Equal(suite.T(), "", results[3]["Received date"], "Received date - client 1")
	assert.Contains(suite.T(), results[3]["Sirius upload date"], runTime.Format("2006-01-02"), "Sirius upload date - client 1")
	assert.Equal(suite.T(), "0", results[3]["Cash amount"], "Cash amount - client 1")
	assert.Equal(suite.T(), "100.00", results[3]["Credit amount"], "Credit amount - client 1")
	assert.Equal(suite.T(), "0", results[3]["Adjustment amount"], "Adjustment amount - client 1")
	assert.Equal(suite.T(), "Test exemption", results[3]["Memo line description"], "Memo line description - client 1")
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
