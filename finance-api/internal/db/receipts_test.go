package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"strconv"
)

func (suite *IntegrationSuite) Test_receipts() {
	ctx := suite.ctx
	today := suite.seeder.Today()
	yesterday := suite.seeder.Today().Sub(0, 0, 1)
	twoYearsAgo := suite.seeder.Today().Sub(2, 0, 0)
	twoMonthsAgo := suite.seeder.Today().Sub(0, 2, 0)
	oneMonthAgo := suite.seeder.Today().Sub(0, 1, 0)
	courtRef := "12345678"

	// transaction timeline:
	// 1st invoice
	// paid in full
	// fee reduction unapplies 50%
	// 2nd invoice
	// 50% paid with reapply
	// 3rd invoice
	// payment covers 2nd and 3rd invoices, with excess
	clientID := suite.seeder.CreateClient(ctx, "Ian", "Test", courtRef, "1234")
	_, inv1Ref := suite.seeder.CreateInvoice(ctx, clientID, shared.InvoiceTypeGA, nil, twoYearsAgo.StringPtr(), nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 20000, twoYearsAgo.Date(), courtRef, shared.TransactionTypeOPGBACSPayment, twoYearsAgo.Date())
	suite.seeder.CreateFeeReduction(ctx, clientID, shared.FeeReductionTypeRemission, strconv.Itoa(twoYearsAgo.Date().Year()-1), 2, "A reduction", twoYearsAgo.Date())

	_, inv2Ref := suite.seeder.CreateInvoice(ctx, clientID, shared.InvoiceTypeS2, valToPtr("316.24"), twoMonthsAgo.StringPtr(), nil, nil, nil)
	_, inv3Ref := suite.seeder.CreateInvoice(ctx, clientID, shared.InvoiceTypeSO, valToPtr("70.00"), twoMonthsAgo.StringPtr(), nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 30000, oneMonthAgo.Date(), courtRef, shared.TransactionTypeMotoCardPayment, oneMonthAgo.Date())

	// excluded as out of range - would have partial reapply if included
	_, _ = suite.seeder.CreateInvoice(ctx, clientID, shared.InvoiceTypeGA, nil, today.StringPtr(), nil, nil, nil)

	c := Client{suite.seeder.Conn}

	from := shared.NewDate(twoYearsAgo.Sub(0, 0, 1).String())
	to := shared.NewDate(yesterday.String())

	rows, err := c.Run(ctx, &Receipts{
		FromDate: &from,
		ToDate:   &to,
	})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 7, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// 1st invoice paid in full
	assert.Equal(suite.T(), "Ian Test", results[0]["Customer name"], "Line 1: Customer name")
	assert.Equal(suite.T(), courtRef, results[0]["Customer number"], "Line 1: Customer number")
	assert.Equal(suite.T(), "1234", results[0]["SOP number"], "Line 1: SOP number")
	assert.Equal(suite.T(), "0470", results[0]["Entity"], "Line 1: Entity")
	assert.Equal(suite.T(), "99999999", results[0]["Receivables cost centre"], "Line 1: Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[0]["Receivables cost centre description"], "Line 1: Receivables cost centre description")
	//assert.Equal(suite.T(), "1816100000", results[0]["Receivables account code"], "Line 1: Receivables account code")
	//assert.Equal(suite.T(), "CA - TRADE RECEIVABLES", results[0]["Account code description"], "Line 1: Account code description")
	assert.Equal(suite.T(), "BC"+inv1Ref, results[0]["Txn number"], "Line 1: Txn number")
	assert.Equal(suite.T(), "BACS Payment", results[0]["Txn type"], "Line 1: Txn type")
	assert.Equal(suite.T(), twoYearsAgo.String(), results[0]["Receipt date"], "Line 1: Receipt date")
	assert.Equal(suite.T(), twoYearsAgo.String(), results[0]["Sirius upload date"], "Line 1: Sirius upload date")
	assert.Equal(suite.T(), twoYearsAgo.FinancialYear(), results[0]["Financial Year"], "Line 1: Financial Year")
	assert.Equal(suite.T(), "200.00", results[0]["Receipt amount"], "Line 1: Receipt amount")
	assert.Equal(suite.T(), "200.00", results[0]["Amount applied"], "Line 1: Amount applied")
	assert.Equal(suite.T(), "0.00", results[0]["Amount unapplied"], "Line 1: Amount unapplied")

	//fee reduction unapplies 50%
	assert.Equal(suite.T(), "Ian Test", results[1]["Customer name"], "Line 2: Customer name")
	assert.Equal(suite.T(), courtRef, results[1]["Customer number"], "Line 2: Customer number")
	assert.Equal(suite.T(), "1234", results[1]["SOP number"], "Line 2: SOP number")
	assert.Equal(suite.T(), "0470", results[1]["Entity"], "Line 2: Entity")
	assert.Equal(suite.T(), "99999999", results[1]["Receivables cost centre"], "Line 2: Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[1]["Receivables cost centre description"], "Line 2: Receivables cost centre description")
	//assert.Equal(suite.T(), "1816100001", results[1]["Receivables account code"], "Line 2: Receivables account code")
	//assert.Equal(suite.T(), "CA - TRADE RECEIVABLES - UNAPPLIED RECEIPTS", results[1]["Account code description"], "Line 2: Account code description")
	assert.Equal(suite.T(), "UA"+inv1Ref, results[1]["Txn number"], "Line 2: Txn number")
	assert.Equal(suite.T(), "Unapply (money from invoice)", results[1]["Txn type"], "Line 2: Txn type")
	assert.Equal(suite.T(), "", results[1]["Receipt date"], "Line 2: Receipt date")
	assert.Equal(suite.T(), twoYearsAgo.String(), results[1]["Sirius upload date"], "Line 2: Sirius upload date")
	assert.Equal(suite.T(), twoYearsAgo.FinancialYear(), results[1]["Financial Year"], "Line 2: Financial Year")
	assert.Equal(suite.T(), "0.00", results[1]["Receipt amount"], "Line 2: Receipt amount")
	assert.Equal(suite.T(), "0.00", results[1]["Amount applied"], "Line 2: Amount applied")
	assert.Equal(suite.T(), "100.00", results[1]["Amount unapplied"], "Line 2: Amount unapplied")

	// 2nd invoice 50% paid with reapply
	assert.Equal(suite.T(), "Ian Test", results[2]["Customer name"], "Line 3: Customer name")
	assert.Equal(suite.T(), courtRef, results[2]["Customer number"], "Line 3: Customer number")
	assert.Equal(suite.T(), "1234", results[2]["SOP number"], "Line 3: SOP number")
	assert.Equal(suite.T(), "0470", results[2]["Entity"], "Line 3: Entity")
	assert.Equal(suite.T(), "99999999", results[2]["Receivables cost centre"], "Line 3: Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[2]["Receivables cost centre description"], "Line 3: Receivables cost centre description")
	//assert.Equal(suite.T(), "1816100000", results[2]["Receivables account code"], "Line 3: Receivables account code")
	//assert.Equal(suite.T(), "CA - TRADE RECEIVABLES", results[2]["Account code description"], "Line 3: Account code description")
	assert.Equal(suite.T(), "RA"+inv2Ref, results[2]["Txn number"], "Line 3: Txn number")
	assert.Equal(suite.T(), "Reapply/Reallocate (money to invoice)", results[2]["Txn type"], "Line 3: Txn type")
	assert.Equal(suite.T(), "", results[2]["Receipt date"], "Line 3: Receipt date")
	assert.Equal(suite.T(), twoMonthsAgo.String(), results[2]["Sirius upload date"], "Line 3: Sirius upload date")
	assert.Equal(suite.T(), twoMonthsAgo.FinancialYear(), results[2]["Financial Year"], "Line 3: Financial Year")
	assert.Equal(suite.T(), "0.00", results[2]["Receipt amount"], "Line 3: Receipt amount")
	assert.Equal(suite.T(), "100.00", results[2]["Amount applied"], "Line 3: Amount applied")
	assert.Equal(suite.T(), "0.00", results[2]["Amount unapplied"], "Line 3: Amount unapplied")

	// payment covers 2nd invoice...
	assert.Equal(suite.T(), "Ian Test", results[3]["Customer name"], "Line 4: Customer name")
	assert.Equal(suite.T(), courtRef, results[3]["Customer number"], "Line 4: Customer number")
	assert.Equal(suite.T(), "1234", results[3]["SOP number"], "Line 4: SOP number")
	assert.Equal(suite.T(), "0470", results[3]["Entity"], "Line 4: Entity")
	assert.Equal(suite.T(), "99999999", results[3]["Receivables cost centre"], "Line 4: Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[3]["Receivables cost centre description"], "Line 4: Receivables cost centre description")
	//assert.Equal(suite.T(), "1816100000", results[3]["Receivables account code"], "Line 4: Receivables account code")
	//assert.Equal(suite.T(), "CA - TRADE RECEIVABLES", results[3]["Account code description"], "Line 4: Account code description")
	assert.Equal(suite.T(), "PC"+inv2Ref, results[3]["Txn number"], "Line 4: Txn number")
	assert.Equal(suite.T(), "MOTO (phone) Card Payment", results[3]["Txn type"], "Line 4: Txn type")
	assert.Equal(suite.T(), oneMonthAgo.String(), results[3]["Receipt date"], "Line 4: Receipt date")
	assert.Equal(suite.T(), oneMonthAgo.String(), results[3]["Sirius upload date"], "Line 4: Sirius upload date")
	assert.Equal(suite.T(), oneMonthAgo.FinancialYear(), results[3]["Financial Year"], "Line 4: Financial Year")
	assert.Equal(suite.T(), "300.00", results[3]["Receipt amount"], "Line 4: Receipt amount")
	assert.Equal(suite.T(), "216.24", results[3]["Amount applied"], "Line 4: Amount applied")
	assert.Equal(suite.T(), "0.00", results[3]["Amount unapplied"], "Line 4: Amount unapplied")

	// ... and 3rd invoice
	assert.Equal(suite.T(), "Ian Test", results[4]["Customer name"], "Line 5: Customer name")
	assert.Equal(suite.T(), courtRef, results[4]["Customer number"], "Line 5: Customer number")
	assert.Equal(suite.T(), "1234", results[4]["SOP number"], "Line 5: SOP number")
	assert.Equal(suite.T(), "0470", results[4]["Entity"], "Line 5: Entity")
	assert.Equal(suite.T(), "99999999", results[4]["Receivables cost centre"], "Line 5: Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[4]["Receivables cost centre description"], "Line 5: Receivables cost centre description")
	//assert.Equal(suite.T(), "1816100000", results[4]["Receivables account code"], "Line 5: Receivables account code")
	//assert.Equal(suite.T(), "CA - TRADE RECEIVABLES", results[4]["Account code description"], "Line 5: Account code description")
	assert.Equal(suite.T(), "PC"+inv3Ref, results[4]["Txn number"], "Line 5: Txn number")
	assert.Equal(suite.T(), "MOTO (phone) Card Payment", results[4]["Txn type"], "Line 5: Txn type")
	assert.Equal(suite.T(), oneMonthAgo.String(), results[4]["Receipt date"], "Line 5: Receipt date")
	assert.Equal(suite.T(), oneMonthAgo.String(), results[4]["Sirius upload date"], "Line 5: Sirius upload date")
	assert.Equal(suite.T(), oneMonthAgo.FinancialYear(), results[4]["Financial Year"], "Line 5: Financial Year")
	assert.Equal(suite.T(), "300.00", results[4]["Receipt amount"], "Line 5: Receipt amount")
	assert.Equal(suite.T(), "70.00", results[4]["Amount applied"], "Line 5: Amount applied")
	assert.Equal(suite.T(), "0.00", results[4]["Amount unapplied"], "Line 5: Amount unapplied")

	// ... and overpays excess
	assert.Equal(suite.T(), "Ian Test", results[5]["Customer name"], "Line 6: Customer name")
	assert.Equal(suite.T(), courtRef, results[5]["Customer number"], "Line 6: Customer number")
	assert.Equal(suite.T(), "1234", results[5]["SOP number"], "Line 6: SOP number")
	assert.Equal(suite.T(), "0470", results[5]["Entity"], "Line 6: Entity")
	assert.Equal(suite.T(), "99999999", results[5]["Receivables cost centre"], "Line 6: Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[5]["Receivables cost centre description"], "Line 6: Receivables cost centre description")
	assert.Equal(suite.T(), "1816100002", results[5]["Receivables account code"], "Line 6: Receivables account code")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES - ON ACCOUNT RECEIPTS", results[5]["Account code description"], "Line 6: Account code description")
	assert.Equal(suite.T(), "PC"+courtRef, results[5]["Txn number"], "Line 6: Txn number")
	assert.Equal(suite.T(), "MOTO (phone) Card Payment", results[5]["Txn type"], "Line 6: Txn type")
	assert.Equal(suite.T(), oneMonthAgo.String(), results[5]["Receipt date"], "Line 6: Receipt date")
	assert.Equal(suite.T(), oneMonthAgo.String(), results[5]["Sirius upload date"], "Line 6: Sirius upload date")
	assert.Equal(suite.T(), oneMonthAgo.FinancialYear(), results[5]["Financial Year"], "Line 6: Financial Year")
	assert.Equal(suite.T(), "300.00", results[5]["Receipt amount"], "Line 6: Receipt amount")
	assert.Equal(suite.T(), "0.00", results[5]["Amount applied"], "Line 6: Amount applied")
	assert.Equal(suite.T(), "13.76", results[5]["Amount unapplied"], "Line 6: Amount unapplied")

	// refunds excess (not yet implemented)
}
