package db

import (
	"strconv"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_paid_invoices() {
	ctx := suite.ctx

	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)
	twoMonthsAgo := today.Sub(0, 2, 0)
	twoYearsAgo := today.Sub(2, 0, 0)
	fourYearsAgo := today.Sub(4, 0, 0)
	general := "320.00"
	minimal := "10.00"

	// client with:
	// one invoice
	// one exemption
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "11111111", "1111", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client1ID, "pfa")
	_, c1i1Ref := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, nil)
	_ = suite.seeder.CreateFeeReduction(ctx, client1ID, shared.FeeReductionTypeExemption, strconv.Itoa(twoYearsAgo.Date().Year()), 3, "Test exemption", today.Sub(0, 0, 3).Date())

	// client with:
	// one invoice with no outstanding balance due to an exemption
	// one invoice with outstanding balance
	client2ID := suite.seeder.CreateClient(ctx, "John", "Suite", "22222222", "2222", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client2ID, "pfa")
	_, c2i1Ref := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeAD, nil, fourYearsAgo.StringPtr(), nil, nil, nil, nil)
	_, _ = suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS2, &general, twoMonthsAgo.StringPtr(), twoMonthsAgo.StringPtr(), nil, nil, nil)
	_ = suite.seeder.CreateFeeReduction(ctx, client2ID, shared.FeeReductionTypeExemption, strconv.Itoa(fourYearsAgo.Date().Year()-1), 2, "Test exemption", today.Sub(0, 0, 1).Date())

	// client with:
	// one invoice partially paid due to a remission
	client3ID := suite.seeder.CreateClient(ctx, "Tony", "Three", "33333333", "3333", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client3ID, "pfa")
	_, _ = suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeAD, nil, fourYearsAgo.StringPtr(), nil, nil, nil, nil)
	_ = suite.seeder.CreateFeeReduction(ctx, client3ID, shared.FeeReductionTypeRemission, strconv.Itoa(fourYearsAgo.Date().Year()-1), 4, "Test remission", today.Date())

	// client with:
	//one invoice paid with supervision BACS payment
	client4ref := "44444444"
	client4ID := suite.seeder.CreateClient(ctx, "Sally", "Supervision", client4ref, "4444", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client4ID, "pfa")
	_, c4i1Ref := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeS3, &minimal, yesterday.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 1000, today.Sub(0, 0, 6).Date(), client4ref, shared.TransactionTypeSupervisionBACSPayment, today.Sub(0, 0, 6).Date(), 0)

	// client with:
	// one invoice paid with OPG BACS payment
	client5ref := "55555555"
	client5ID := suite.seeder.CreateClient(ctx, "Owen", "OPG", client5ref, "5555", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client5ID, "pfa")
	_, c5i1Ref := suite.seeder.CreateInvoice(ctx, client5ID, shared.InvoiceTypeS2, &general, today.Sub(0, 0, 4).StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 32000, today.Sub(0, 0, 4).Date(), client5ref, shared.TransactionTypeOPGBACSPayment, today.Sub(0, 0, 4).Date(), 0)

	// client with:
	// one Guardianship invoice paid with OPG BACS payment and remission
	client6ref := "66666666"
	client6ID := suite.seeder.CreateClient(ctx, "Gary", "Guardianship", client6ref, "6666", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client6ID, "pfa")
	_, c6i1Ref := suite.seeder.CreateInvoice(ctx, client6ID, shared.InvoiceTypeGA, valToPtr("200.00"), today.Sub(0, 0, 2).StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 10000, today.Sub(0, 0, 2).Date(), client6ref, shared.TransactionTypeOPGBACSPayment, today.Sub(0, 0, 2).Date(), 0)
	_ = suite.seeder.CreateFeeReduction(ctx, client6ID, shared.FeeReductionTypeRemission, strconv.Itoa(today.Sub(0, 0, 2).Date().Year()-1), 2, "Gary's remission", today.Sub(0, 0, 2).Date())

	// client with:
	// one Guardianship invoice with exemption
	client7ref := "77777777"
	client7ID := suite.seeder.CreateClient(ctx, "Edith", "Exemption", client7ref, "7777", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client7ID, "pfa")
	_, c7i1Ref := suite.seeder.CreateInvoice(ctx, client7ID, shared.InvoiceTypeGS, valToPtr("200.00"), today.StringPtr(), today.StringPtr(), today.StringPtr(), nil, nil)
	_ = suite.seeder.CreateFeeReduction(ctx, client7ID, shared.FeeReductionTypeExemption, strconv.Itoa(today.Date().Year()-1), 2, "Edith's exemption", today.Sub(0, 0, 5).Date())

	// misapplied payments
	// first client will not show in report as the payment has been reversed (and the invoice is not paid)
	client8ref := "88888888"
	client8ID := suite.seeder.CreateClient(ctx, "Ernie", "Error", client8ref, "2222", "ACTIVE")
	_, _ = suite.seeder.CreateInvoice(ctx, client8ID, shared.InvoiceTypeAD, nil, today.Sub(0, 0, 7).StringPtr(), nil, nil, nil, today.Sub(0, 0, 7).StringPtr())
	suite.seeder.CreatePayment(ctx, 15000, today.Sub(0, 0, 7).Date(), client8ref, shared.TransactionTypeOnlineCardPayment, today.Sub(0, 0, 7).Date(), 0)
	client9ref := "99999999"
	client9ID := suite.seeder.CreateClient(ctx, "Colette", "Correct", client9ref, "3333", "ACTIVE")
	_, c9i1Ref := suite.seeder.CreateInvoice(ctx, client9ID, shared.InvoiceTypeSO, valToPtr("90.00"), today.Sub(0, 0, 7).StringPtr(), nil, nil, nil, today.Sub(0, 0, 7).StringPtr())
	suite.seeder.ReversePayment(ctx, client8ref, client9ref, "150.00", today.Sub(0, 0, 7).Date(), today.Sub(0, 0, 7).Date(), shared.TransactionTypeOnlineCardPayment, today.Date(), "")

	c := Client{suite.seeder.Conn}

	from := shared.NewDate(fourYearsAgo.String())
	to := shared.NewDate(today.String())

	rows, err := c.Run(ctx, NewPaidInvoices(PaidInvoicesInput{
		FromDate: &from,
		ToDate:   &to,
	}))

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 9, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// misapplied payment
	assert.Equal(suite.T(), "Colette Correct", results[0]["Customer name"], "Customer Name - client 9")
	assert.Equal(suite.T(), client9ref, results[0]["Customer number"], "Customer number - client 9")
	assert.Equal(suite.T(), "3333", results[0]["SOP number"], "SOP number - client 9")
	assert.Equal(suite.T(), "=\"0470\"", results[0]["Entity"], "Entity - client 9")
	assert.Equal(suite.T(), "99999999", results[0]["Cost centre"], "Cost centre - client 9")
	assert.Equal(suite.T(), "BALANCE SHEET", results[0]["Cost centre description"], "Cost centre description - client 9")
	assert.Equal(suite.T(), "1816102003", results[0]["Account code"], "Account code - client 9")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES - SIRIUS SUPERVISION CONTROL ACCOUNT", results[0]["Account code description"], "Account code description - client 9")
	assert.Equal(suite.T(), "SO", results[0]["Invoice type"], "Invoice type - client 9")
	assert.Equal(suite.T(), c9i1Ref, results[0]["Invoice number"], "Invoice number - client 9")
	assert.Equal(suite.T(), "OC"+c9i1Ref, results[0]["Txn number"], "Txn number - client 9")
	assert.Equal(suite.T(), "Online Card Payment", results[0]["Txn description"], "Txn description - client 9")
	assert.Equal(suite.T(), "90.00", results[0]["Original amount"], "Original amount - client 9")
	assert.Equal(suite.T(), today.Sub(0, 0, 7).String(), results[0]["Received date"], "Received date - client 9")
	assert.Contains(suite.T(), today.Sub(0, 0, 7).String(), results[0]["Sirius upload date"], "Sirius upload date - client 9")
	assert.Equal(suite.T(), "90.00", results[0]["Cash amount"], "Cash amount - client 9")
	assert.Equal(suite.T(), "0", results[0]["Credit amount"], "Credit amount - client 9")
	assert.Equal(suite.T(), "0", results[0]["Adjustment amount"], "Adjustment amount - client 9")
	assert.Equal(suite.T(), "", results[0]["Memo line description"], "Memo line description - client 9")

	// client 4
	assert.Equal(suite.T(), "Sally Supervision", results[1]["Customer name"], "Customer Name - client 4")
	assert.Equal(suite.T(), client4ref, results[1]["Customer number"], "Customer number - client 4")
	assert.Equal(suite.T(), "4444", results[1]["SOP number"], "SOP number - client 4")
	assert.Equal(suite.T(), "=\"0470\"", results[1]["Entity"], "Entity - client 4")
	assert.Equal(suite.T(), "99999999", results[1]["Cost centre"], "Cost centre - client 4")
	assert.Equal(suite.T(), "BALANCE SHEET", results[1]["Cost centre description"], "Cost centre description - client 4")
	assert.Equal(suite.T(), "1816102003", results[1]["Account code"], "Account code - client 4")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES - SIRIUS SUPERVISION CONTROL ACCOUNT", results[1]["Account code description"], "Account code description - client 4")
	assert.Equal(suite.T(), "S3", results[1]["Invoice type"], "Invoice type - client 4")
	assert.Equal(suite.T(), c4i1Ref, results[1]["Invoice number"], "Invoice number - client 4")
	assert.Equal(suite.T(), "BC"+c4i1Ref, results[1]["Txn number"], "Txn number - client 4")
	assert.Equal(suite.T(), "BACS Payment", results[1]["Txn description"], "Txn description - client 4")
	assert.Equal(suite.T(), "10.00", results[1]["Original amount"], "Original amount - client 4")
	assert.Equal(suite.T(), today.Sub(0, 0, 6).String(), results[1]["Received date"], "Received date - client 4")
	assert.Contains(suite.T(), today.Sub(0, 0, 6).String(), results[1]["Sirius upload date"], "Sirius upload date - client 4")
	assert.Equal(suite.T(), "10.00", results[1]["Cash amount"], "Cash amount - client 4")
	assert.Equal(suite.T(), "0", results[1]["Credit amount"], "Credit amount - client 4")
	assert.Equal(suite.T(), "0", results[1]["Adjustment amount"], "Adjustment amount - client 4")
	assert.Equal(suite.T(), "", results[1]["Memo line description"], "Memo line description - client 4")

	// client 7 - remission
	assert.Equal(suite.T(), "Edith Exemption", results[2]["Customer name"], "Customer Name - client 7 - exemption")
	assert.Equal(suite.T(), "77777777", results[2]["Customer number"], "Customer number - client 7 - exemption")
	assert.Equal(suite.T(), "7777", results[2]["SOP number"], "SOP number - client 7 - exemption")
	assert.Equal(suite.T(), "=\"0470\"", results[2]["Entity"], "Entity - client 7 - exemption")
	assert.Equal(suite.T(), "10486000", results[2]["Cost centre"], "Cost centre - client 7 - exemption")
	assert.Equal(suite.T(), "Allocations, HW & SIS BISD", results[2]["Cost centre description"], "Cost centre description - client 7 - exemption")
	assert.Equal(suite.T(), "4481102108", results[2]["Account code"], "Account code - client 7 - exemption")
	assert.Equal(suite.T(), "INC - RECEIPT OF FEES AND CHARGES - GUARDIANSHIP FEE EXEMPTION", results[2]["Account code description"], "Account code description - client 7 - exemption")
	assert.Equal(suite.T(), "GS", results[2]["Invoice type"], "Invoice type - client 7 - exemption")
	assert.Equal(suite.T(), c7i1Ref, results[2]["Invoice number"], "Invoice number - client 7 - exemption")
	assert.Equal(suite.T(), "ZE"+c7i1Ref, results[2]["Txn number"], "Txn number - client 7 - exemption")
	assert.Equal(suite.T(), "Exemption Credit", results[2]["Txn description"], "Txn description - client 7 - exemption")
	assert.Equal(suite.T(), "200.00", results[2]["Original amount"], "Original amount - client 7 - exemption")
	assert.Equal(suite.T(), "", results[2]["Received date"], "Received date - client 7 - exemption")
	assert.Contains(suite.T(), today.Sub(0, 0, 5).String(), results[2]["Sirius upload date"], "Sirius upload date - client 7 - exemption")
	assert.Equal(suite.T(), "0", results[2]["Cash amount"], "Cash amount - client 7 - exemption")
	assert.Equal(suite.T(), "200.00", results[2]["Credit amount"], "Credit amount - client 7 - exemption")
	assert.Equal(suite.T(), "0", results[2]["Adjustment amount"], "Adjustment amount - client 7 - exemption")
	assert.Equal(suite.T(), "Edith's exemption", results[2]["Memo line description"], "Memo line description - client 7 - exemption")

	// client 5
	assert.Equal(suite.T(), "Owen OPG", results[3]["Customer name"], "Customer Name - client 5")
	assert.Equal(suite.T(), client5ref, results[3]["Customer number"], "Customer number - client 5")
	assert.Equal(suite.T(), "5555", results[3]["SOP number"], "SOP number - client 5")
	assert.Equal(suite.T(), "=\"0470\"", results[3]["Entity"], "Entity - client 5")
	assert.Equal(suite.T(), "99999999", results[3]["Cost centre"], "Cost centre - client 5")
	assert.Equal(suite.T(), "BALANCE SHEET", results[3]["Cost centre description"], "Cost centre description - client 5")
	assert.Equal(suite.T(), "1816102003", results[3]["Account code"], "Account code - client 5")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES - SIRIUS SUPERVISION CONTROL ACCOUNT", results[3]["Account code description"], "Account code description - client 5")
	assert.Equal(suite.T(), "S2", results[3]["Invoice type"], "Invoice type - client 5")
	assert.Equal(suite.T(), c5i1Ref, results[3]["Invoice number"], "Invoice number - client 5")
	assert.Equal(suite.T(), "BC"+c5i1Ref, results[3]["Txn number"], "Txn number - client 5")
	assert.Equal(suite.T(), "BACS Payment", results[3]["Txn description"], "Txn description - client 5")
	assert.Equal(suite.T(), "320.00", results[3]["Original amount"], "Original amount - client 5")
	assert.Equal(suite.T(), today.Sub(0, 0, 4).String(), results[3]["Received date"], "Received date - client 5")
	assert.Contains(suite.T(), today.Sub(0, 0, 4).String(), results[3]["Sirius upload date"], "Sirius upload date - client 5")
	assert.Equal(suite.T(), "320.00", results[3]["Cash amount"], "Cash amount - client 5")
	assert.Equal(suite.T(), "0", results[3]["Credit amount"], "Credit amount - client 5")
	assert.Equal(suite.T(), "0", results[3]["Adjustment amount"], "Adjustment amount - client 5")
	assert.Equal(suite.T(), "", results[3]["Memo line description"], "Memo line description - client 5")

	// client 1
	assert.Equal(suite.T(), "Ian Test", results[4]["Customer name"], "Customer Name - client 1")
	assert.Equal(suite.T(), "11111111", results[4]["Customer number"], "Customer number - client 1")
	assert.Equal(suite.T(), "1111", results[4]["SOP number"], "SOP number - client 1")
	assert.Equal(suite.T(), "=\"0470\"", results[4]["Entity"], "Entity - client 1")
	assert.Equal(suite.T(), "10482009", results[4]["Cost centre"], "Cost centre - client 1")
	assert.Equal(suite.T(), "Supervision Investigations", results[4]["Cost centre description"], "Cost centre description - client 1")
	assert.Equal(suite.T(), "4481102114", results[4]["Account code"], "Account code - client 1")
	assert.Equal(suite.T(), "INC - RECEIPT OF FEES AND CHARGES - Rem Appoint Deputy", results[4]["Account code description"], "Account code description - client 1")
	assert.Equal(suite.T(), "AD", results[4]["Invoice type"], "Invoice type - client 1")
	assert.Equal(suite.T(), c1i1Ref, results[4]["Invoice number"], "Invoice number - client 1")
	assert.Equal(suite.T(), "ZE"+c1i1Ref, results[4]["Txn number"], "Txn number - client 1")
	assert.Equal(suite.T(), "Exemption Credit", results[4]["Txn description"], "Txn description - client 1")
	assert.Equal(suite.T(), "100.00", results[4]["Original amount"], "Original amount - client 1")
	assert.Equal(suite.T(), "", results[4]["Received date"], "Received date - client 1")
	assert.Contains(suite.T(), today.Sub(0, 0, 3).String(), results[4]["Sirius upload date"], "Sirius upload date - client 1")
	assert.Equal(suite.T(), "0", results[4]["Cash amount"], "Cash amount - client 1")
	assert.Equal(suite.T(), "100.00", results[4]["Credit amount"], "Credit amount - client 1")
	assert.Equal(suite.T(), "0", results[4]["Adjustment amount"], "Adjustment amount - client 1")
	assert.Equal(suite.T(), "Test exemption", results[4]["Memo line description"], "Memo line description - client 1")

	// client 6 - payment
	assert.Equal(suite.T(), "Gary Guardianship", results[5]["Customer name"], "Customer Name - client 6 - payment")
	assert.Equal(suite.T(), client6ref, results[5]["Customer number"], "Customer number - client 6 - payment")
	assert.Equal(suite.T(), "6666", results[5]["SOP number"], "SOP number - client 6 - payment")
	assert.Equal(suite.T(), "=\"0470\"", results[5]["Entity"], "Entity - client 6 - payment")
	assert.Equal(suite.T(), "99999999", results[5]["Cost centre"], "Cost centre - client 6 - payment")
	assert.Equal(suite.T(), "BALANCE SHEET", results[5]["Cost centre description"], "Cost centre description - client 6 - payment")
	assert.Equal(suite.T(), "1816102003", results[5]["Account code"], "Account code - client 6 - payment")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES - SIRIUS SUPERVISION CONTROL ACCOUNT", results[5]["Account code description"], "Account code description - client 6 - payment")
	assert.Equal(suite.T(), "GA", results[5]["Invoice type"], "Invoice type - client 6 - payment")
	assert.Equal(suite.T(), c6i1Ref, results[5]["Invoice number"], "Invoice number - client 6 - payment")
	assert.Equal(suite.T(), "BC"+c6i1Ref, results[5]["Txn number"], "Txn number - client 6 - payment")
	assert.Equal(suite.T(), "BACS Payment", results[5]["Txn description"], "Txn description - client 6 - payment")
	assert.Equal(suite.T(), "200.00", results[5]["Original amount"], "Original amount - client 6 - payment")
	assert.Equal(suite.T(), today.Sub(0, 0, 2).String(), results[5]["Received date"], "Received date - client 6 - payment")
	assert.Contains(suite.T(), today.Sub(0, 0, 2).String(), results[5]["Sirius upload date"], "Sirius upload date - client 6 - payment")
	assert.Equal(suite.T(), "100.00", results[5]["Cash amount"], "Cash amount - client 6 - payment")
	assert.Equal(suite.T(), "0", results[5]["Credit amount"], "Credit amount - client 6 - payment")
	assert.Equal(suite.T(), "0", results[5]["Adjustment amount"], "Adjustment amount - client 6 - payment")
	assert.Equal(suite.T(), "", results[5]["Memo line description"], "Memo line description - client 6 - payment")

	// client 6 - remission
	assert.Equal(suite.T(), "Gary Guardianship", results[6]["Customer name"], "Customer Name - client6 - remission")
	assert.Equal(suite.T(), client6ref, results[6]["Customer number"], "Customer number - client6 - remission")
	assert.Equal(suite.T(), "6666", results[6]["SOP number"], "SOP number - client6 - remission")
	assert.Equal(suite.T(), "=\"0470\"", results[6]["Entity"], "Entity - client6 - remission")
	assert.Equal(suite.T(), "10486000", results[6]["Cost centre"], "Cost centre - client6 - remission")
	assert.Equal(suite.T(), "Allocations, HW & SIS BISD", results[6]["Cost centre description"], "Cost centre description - client6 - remission")
	assert.Equal(suite.T(), "4481102107", results[6]["Account code"], "Account code - client6 - remission")
	assert.Equal(suite.T(), "INC - RECEIPT OF FEES AND CHARGES - GUARDIANSHIP FEE REMISSION", results[6]["Account code description"], "Account code description - client6 - remission")
	assert.Equal(suite.T(), "GA", results[6]["Invoice type"], "Invoice type - client6 - remission")
	assert.Equal(suite.T(), c6i1Ref, results[6]["Invoice number"], "Invoice number - client6 - remission")
	assert.Equal(suite.T(), "ZR"+c6i1Ref, results[6]["Txn number"], "Txn number - client6 - remission")
	assert.Equal(suite.T(), "Remission Credit", results[6]["Txn description"], "Txn description - client6 - remission")
	assert.Equal(suite.T(), "200.00", results[6]["Original amount"], "Original amount - client6 - remission")
	assert.Equal(suite.T(), "", results[6]["Received date"], "Received date - client6 - remission")
	assert.Contains(suite.T(), today.Sub(0, 0, 2).String(), results[6]["Sirius upload date"], "Sirius upload date - client6 - remission")
	assert.Equal(suite.T(), "0", results[6]["Cash amount"], "Cash amount - client6 - remission")
	assert.Equal(suite.T(), "100.00", results[6]["Credit amount"], "Credit amount - client6 - remission")
	assert.Equal(suite.T(), "0", results[6]["Adjustment amount"], "Adjustment amount - client6 - remission")
	assert.Equal(suite.T(), "Gary's remission", results[6]["Memo line description"], "Memo line description - client6 - remission")

	// client 2 invoice 1
	assert.Equal(suite.T(), "John Suite", results[7]["Customer name"], "Customer Name - client 2 invoice 1")
	assert.Equal(suite.T(), "22222222", results[7]["Customer number"], "Customer number - client 2 invoice 1")
	assert.Equal(suite.T(), "2222", results[7]["SOP number"], "SOP number - client 2 invoice 1")
	assert.Equal(suite.T(), "=\"0470\"", results[7]["Entity"], "Entity - client 2 invoice 1")
	assert.Equal(suite.T(), "10482009", results[7]["Cost centre"], "Cost centre - client 2 invoice 1")
	assert.Equal(suite.T(), "Supervision Investigations", results[7]["Cost centre description"], "Cost centre description - client 2 invoice 1")
	assert.Equal(suite.T(), "4481102114", results[7]["Account code"], "Account code - client 2 invoice 1")
	assert.Equal(suite.T(), "INC - RECEIPT OF FEES AND CHARGES - Rem Appoint Deputy", results[7]["Account code description"], "Account code description - client 2 invoice 1")
	assert.Equal(suite.T(), "AD", results[7]["Invoice type"], "Invoice type - client 2 invoice 1")
	assert.Equal(suite.T(), c2i1Ref, results[7]["Invoice number"], "Invoice number - client 2 invoice 1")
	assert.Equal(suite.T(), "ZE"+c2i1Ref, results[7]["Txn number"], "Txn number - client 2 invoice 1")
	assert.Equal(suite.T(), "Exemption Credit", results[7]["Txn description"], "Txn description - client 2 invoice 1")
	assert.Equal(suite.T(), "100.00", results[7]["Original amount"], "Original amount - client 2 invoice 1")
	assert.Equal(suite.T(), "", results[7]["Received date"], "Received date - client 2 invoice 1")
	assert.Contains(suite.T(), today.Sub(0, 0, 1).String(), results[7]["Sirius upload date"], "Sirius upload date - client 2 invoice 1")
	assert.Equal(suite.T(), "0", results[7]["Cash amount"], "Cash amount - client 2 invoice 1")
	assert.Equal(suite.T(), "100.00", results[7]["Credit amount"], "Credit amount - client 2 invoice 1")
	assert.Equal(suite.T(), "0", results[7]["Adjustment amount"], "Adjustment amount - client 2 invoice 1")
	assert.Equal(suite.T(), "Test exemption", results[7]["Memo line description"], "Memo line description - client 2 invoice 1")
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
			paidInvoices := NewPaidInvoices(PaidInvoicesInput{
				FromDate:   tt.fromDate,
				ToDate:     tt.toDate,
				GoLiveDate: goLiveDate,
			})

			params := paidInvoices.GetParams()
			assert.Equal(t, tt.expected, params)
		})
	}
}
