package db

import (
	"strconv"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_receipts() {
	ctx := suite.ctx
	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)
	twoYearsAgo := today.Sub(2, 0, 0)
	twoMonthsAgo := today.Sub(0, 2, 0)
	oneMonthAgo := today.Sub(0, 1, 0)
	courtRef1 := "12345678"
	courtRef2 := "22222222"
	courtRef3 := "33333333"

	// transaction timeline:
	// 1st invoice
	// paid in full
	// fee reduction unapplies 50%
	// 2nd invoice
	// 50% paid with reapply
	// 3rd invoice
	// payment covers 2nd and 3rd invoices, with excess
	clientID := suite.seeder.CreateClient(ctx, "Ian", "Test", courtRef1, "1234", "ACTIVE")
	_, inv1Ref := suite.seeder.CreateInvoice(ctx, clientID, shared.InvoiceTypeGA, nil, twoYearsAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 20000, twoYearsAgo.Date(), courtRef1, shared.TransactionTypeOPGBACSPayment, twoYearsAgo.Date(), 0)
	_ = suite.seeder.CreateFeeReduction(ctx, clientID, shared.FeeReductionTypeRemission, strconv.Itoa(twoYearsAgo.Date().Year()-1), 2, "A reduction", twoYearsAgo.Date())

	_, inv2Ref := suite.seeder.CreateInvoice(ctx, clientID, shared.InvoiceTypeS2, valToPtr("316.24"), twoMonthsAgo.StringPtr(), nil, nil, nil, nil)
	_, inv3Ref := suite.seeder.CreateInvoice(ctx, clientID, shared.InvoiceTypeSO, valToPtr("70.00"), twoMonthsAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 30000, oneMonthAgo.Date(), courtRef1, shared.TransactionTypeMotoCardPayment, oneMonthAgo.Date(), 0)

	// misapplied payments with overpayment
	client2ID := suite.seeder.CreateClient(ctx, "Ernie", "Error", courtRef2, "2222", "ACTIVE")
	_, inv4Ref := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())
	suite.seeder.CreatePayment(ctx, 15000, yesterday.Date(), courtRef2, shared.TransactionTypeOnlineCardPayment, yesterday.Date(), 0)
	client3ID := suite.seeder.CreateClient(ctx, "Colette", "Correct", courtRef3, "3333", "ACTIVE")
	_, inv5Ref := suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeSO, valToPtr("90.00"), yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())
	suite.seeder.ReversePayment(ctx, courtRef2, courtRef3, "150.00", yesterday.Date(), yesterday.Date(), shared.TransactionTypeOnlineCardPayment, yesterday.Date(), "")

	// excluded as out of range - would have partial reapply if included
	_, _ = suite.seeder.CreateInvoice(ctx, clientID, shared.InvoiceTypeGA, nil, today.StringPtr(), nil, nil, nil, nil)

	c := Client{suite.seeder.Conn}

	from := shared.NewDate(twoYearsAgo.Sub(0, 0, 1).String())
	to := shared.NewDate(yesterday.String())

	rows, err := c.Run(ctx, NewReceipts(ReceiptsInput{
		FromDate: &from,
		ToDate:   &to,
	}))

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 13, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// 1st invoice paid in full
	assert.Equal(suite.T(), "Ian Test", results[0]["Customer name"], "Line 1: Customer name")
	assert.Equal(suite.T(), courtRef1, results[0]["Customer number"], "Line 1: Customer number")
	assert.Equal(suite.T(), "1234", results[0]["SOP number"], "Line 1: SOP number")
	assert.Equal(suite.T(), "0470", results[0]["Entity"], "Line 1: Entity")
	assert.Equal(suite.T(), "99999999", results[0]["Receivables cost centre"], "Line 1: Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[0]["Receivables cost centre description"], "Line 1: Receivables cost centre description")
	assert.Equal(suite.T(), "1816102003", results[0]["Receivables account code"], "Line 1: Receivables account code")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES - SIRIUS SUPERVISION CONTROL ACCOUNT", results[0]["Account code description"], "Line 1: Account code description")
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
	assert.Equal(suite.T(), courtRef1, results[1]["Customer number"], "Line 2: Customer number")
	assert.Equal(suite.T(), "1234", results[1]["SOP number"], "Line 2: SOP number")
	assert.Equal(suite.T(), "0470", results[1]["Entity"], "Line 2: Entity")
	assert.Equal(suite.T(), "99999999", results[1]["Receivables cost centre"], "Line 2: Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[1]["Receivables cost centre description"], "Line 2: Receivables cost centre description")
	assert.Equal(suite.T(), "1816102003", results[1]["Receivables account code"], "Line 2: Receivables account code")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES - SIRIUS SUPERVISION CONTROL ACCOUNT", results[1]["Account code description"], "Line 2: Account code description")
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
	assert.Equal(suite.T(), courtRef1, results[2]["Customer number"], "Line 3: Customer number")
	assert.Equal(suite.T(), "1234", results[2]["SOP number"], "Line 3: SOP number")
	assert.Equal(suite.T(), "0470", results[2]["Entity"], "Line 3: Entity")
	assert.Equal(suite.T(), "99999999", results[2]["Receivables cost centre"], "Line 3: Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[2]["Receivables cost centre description"], "Line 3: Receivables cost centre description")
	assert.Equal(suite.T(), "1816102003", results[2]["Receivables account code"], "Line 3: Receivables account code")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES - SIRIUS SUPERVISION CONTROL ACCOUNT", results[2]["Account code description"], "Line 3: Account code description")
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
	assert.Equal(suite.T(), courtRef1, results[3]["Customer number"], "Line 4: Customer number")
	assert.Equal(suite.T(), "1234", results[3]["SOP number"], "Line 4: SOP number")
	assert.Equal(suite.T(), "0470", results[3]["Entity"], "Line 4: Entity")
	assert.Equal(suite.T(), "99999999", results[3]["Receivables cost centre"], "Line 4: Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[3]["Receivables cost centre description"], "Line 4: Receivables cost centre description")
	assert.Equal(suite.T(), "1816102003", results[3]["Receivables account code"], "Line 4: Receivables account code")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES - SIRIUS SUPERVISION CONTROL ACCOUNT", results[3]["Account code description"], "Line 4: Account code description")
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
	assert.Equal(suite.T(), courtRef1, results[4]["Customer number"], "Line 5: Customer number")
	assert.Equal(suite.T(), "1234", results[4]["SOP number"], "Line 5: SOP number")
	assert.Equal(suite.T(), "0470", results[4]["Entity"], "Line 5: Entity")
	assert.Equal(suite.T(), "99999999", results[4]["Receivables cost centre"], "Line 5: Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[4]["Receivables cost centre description"], "Line 5: Receivables cost centre description")
	assert.Equal(suite.T(), "1816102003", results[4]["Receivables account code"], "Line 5: Receivables account code")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES - SIRIUS SUPERVISION CONTROL ACCOUNT", results[4]["Account code description"], "Line 5: Account code description")
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
	assert.Equal(suite.T(), courtRef1, results[5]["Customer number"], "Line 6: Customer number")
	assert.Equal(suite.T(), "1234", results[5]["SOP number"], "Line 6: SOP number")
	assert.Equal(suite.T(), "0470", results[5]["Entity"], "Line 6: Entity")
	assert.Equal(suite.T(), "99999999", results[5]["Receivables cost centre"], "Line 6: Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[5]["Receivables cost centre description"], "Line 6: Receivables cost centre description")
	assert.Equal(suite.T(), "1816102005", results[5]["Receivables account code"], "Line 6: Receivables account code")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES - ON ACCOUNT RECEIPTS – SIRIUS SUPERVISION", results[5]["Account code description"], "Line 6: Account code description")
	assert.Equal(suite.T(), "PC"+courtRef1, results[5]["Txn number"], "Line 6: Txn number")
	assert.Equal(suite.T(), "MOTO (phone) Card Payment", results[5]["Txn type"], "Line 6: Txn type")
	assert.Equal(suite.T(), oneMonthAgo.String(), results[5]["Receipt date"], "Line 6: Receipt date")
	assert.Equal(suite.T(), oneMonthAgo.String(), results[5]["Sirius upload date"], "Line 6: Sirius upload date")
	assert.Equal(suite.T(), oneMonthAgo.FinancialYear(), results[5]["Financial Year"], "Line 6: Financial Year")
	assert.Equal(suite.T(), "300.00", results[5]["Receipt amount"], "Line 6: Receipt amount")
	assert.Equal(suite.T(), "0.00", results[5]["Amount applied"], "Line 6: Amount applied")
	assert.Equal(suite.T(), "13.76", results[5]["Amount unapplied"], "Line 6: Amount unapplied")

	// misapplied payments with overpayment:

	// original payment
	assert.Equal(suite.T(), "Ernie Error", results[6]["Customer name"], "Line 7: Customer name")
	assert.Equal(suite.T(), courtRef2, results[6]["Customer number"], "Line 7: Customer number")
	assert.Equal(suite.T(), "2222", results[6]["SOP number"], "Line 7: SOP number")
	assert.Equal(suite.T(), "0470", results[6]["Entity"], "Line 7: Entity")
	assert.Equal(suite.T(), "99999999", results[6]["Receivables cost centre"], "Line 7: Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[6]["Receivables cost centre description"], "Line 7: Receivables cost centre description")
	assert.Equal(suite.T(), "1816102003", results[6]["Receivables account code"], "Line 7: Receivables account code")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES - SIRIUS SUPERVISION CONTROL ACCOUNT", results[6]["Account code description"], "Line 7: Account code description")
	assert.Equal(suite.T(), "OC"+inv4Ref, results[6]["Txn number"], "Line 7: Txn number")
	assert.Equal(suite.T(), "Online Card Payment", results[6]["Txn type"], "Line 7: Txn type")
	assert.Equal(suite.T(), yesterday.String(), results[6]["Receipt date"], "Line 7: Receipt date")
	assert.Equal(suite.T(), yesterday.String(), results[6]["Sirius upload date"], "Line 7: Sirius upload date")
	assert.Equal(suite.T(), yesterday.FinancialYear(), results[6]["Financial Year"], "Line 7: Financial Year")
	assert.Equal(suite.T(), "150.00", results[6]["Receipt amount"], "Line 7: Receipt amount")
	assert.Equal(suite.T(), "100.00", results[6]["Amount applied"], "Line 7: Amount applied")
	assert.Equal(suite.T(), "0.00", results[6]["Amount unapplied"], "Line 7: Amount unapplied")

	// and overpayment
	assert.Equal(suite.T(), "Ernie Error", results[7]["Customer name"], "Line 8: Customer name")
	assert.Equal(suite.T(), courtRef2, results[7]["Customer number"], "Line 8: Customer number")
	assert.Equal(suite.T(), "2222", results[7]["SOP number"], "Line 8: SOP number")
	assert.Equal(suite.T(), "0470", results[7]["Entity"], "Line 8: Entity")
	assert.Equal(suite.T(), "99999999", results[7]["Receivables cost centre"], "Line 8: Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[7]["Receivables cost centre description"], "Line 8: Receivables cost centre description")
	assert.Equal(suite.T(), "1816102005", results[7]["Receivables account code"], "Line 8: Receivables account code")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES - ON ACCOUNT RECEIPTS – SIRIUS SUPERVISION", results[7]["Account code description"], "Line 8: Account code description")
	assert.Equal(suite.T(), "OC"+courtRef2, results[7]["Txn number"], "Line 8: Txn number")
	assert.Equal(suite.T(), "Online Card Payment", results[7]["Txn type"], "Line 8: Txn type")
	assert.Equal(suite.T(), yesterday.String(), results[7]["Receipt date"], "Line 8: Receipt date")
	assert.Equal(suite.T(), yesterday.String(), results[7]["Sirius upload date"], "Line 8: Sirius upload date")
	assert.Equal(suite.T(), yesterday.FinancialYear(), results[7]["Financial Year"], "Line 8: Financial Year")
	assert.Equal(suite.T(), "150.00", results[7]["Receipt amount"], "Line 8: Receipt amount")
	assert.Equal(suite.T(), "0.00", results[7]["Amount applied"], "Line 8: Amount applied")
	assert.Equal(suite.T(), "50.00", results[7]["Amount unapplied"], "Line 8: Amount unapplied")

	// reversed
	assert.Equal(suite.T(), "Ernie Error", results[8]["Customer name"], "Line 9: Customer name")
	assert.Equal(suite.T(), courtRef2, results[8]["Customer number"], "Line 9: Customer number")
	assert.Equal(suite.T(), "2222", results[8]["SOP number"], "Line 9: SOP number")
	assert.Equal(suite.T(), "0470", results[8]["Entity"], "Line 9: Entity")
	assert.Equal(suite.T(), "99999999", results[8]["Receivables cost centre"], "Line 9: Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[8]["Receivables cost centre description"], "Line 9: Receivables cost centre description")
	assert.Equal(suite.T(), "1816102005", results[8]["Receivables account code"], "Line 9: Receivables account code")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES - ON ACCOUNT RECEIPTS – SIRIUS SUPERVISION", results[8]["Account code description"], "Line 9: Account code description")
	assert.Equal(suite.T(), "OC"+courtRef2, results[8]["Txn number"], "Line 9: Txn number")
	assert.Equal(suite.T(), "Online Card Payment", results[8]["Txn type"], "Line 9: Txn type")
	assert.Equal(suite.T(), yesterday.String(), results[8]["Receipt date"], "Line 9: Receipt date")
	assert.Equal(suite.T(), yesterday.String(), results[8]["Sirius upload date"], "Line 9: Sirius upload date")
	assert.Equal(suite.T(), yesterday.FinancialYear(), results[8]["Financial Year"], "Line 9: Financial Year")
	assert.Equal(suite.T(), "-150.00", results[8]["Receipt amount"], "Line 9: Receipt amount")
	assert.Equal(suite.T(), "0.00", results[8]["Amount applied"], "Line 9: Amount applied")
	assert.Equal(suite.T(), "-50.00", results[8]["Amount unapplied"], "Line 9: Amount unapplied")

	assert.Equal(suite.T(), "Ernie Error", results[9]["Customer name"], "Line 10: Customer name")
	assert.Equal(suite.T(), courtRef2, results[9]["Customer number"], "Line 10: Customer number")
	assert.Equal(suite.T(), "2222", results[9]["SOP number"], "Line 10: SOP number")
	assert.Equal(suite.T(), "0470", results[9]["Entity"], "Line 10: Entity")
	assert.Equal(suite.T(), "99999999", results[9]["Receivables cost centre"], "Line 10: Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[9]["Receivables cost centre description"], "Line 10: Receivables cost centre description")
	assert.Equal(suite.T(), "1816102003", results[9]["Receivables account code"], "Line 10: Receivables account code")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES - SIRIUS SUPERVISION CONTROL ACCOUNT", results[9]["Account code description"], "Line 10: Account code description")
	assert.Equal(suite.T(), "OC"+inv4Ref, results[9]["Txn number"], "Line 10: Txn number")
	assert.Equal(suite.T(), "Online Card Payment", results[9]["Txn type"], "Line 10: Txn type")
	assert.Equal(suite.T(), yesterday.String(), results[9]["Receipt date"], "Line 10: Receipt date")
	assert.Equal(suite.T(), yesterday.String(), results[9]["Sirius upload date"], "Line 10: Sirius upload date")
	assert.Equal(suite.T(), yesterday.FinancialYear(), results[9]["Financial Year"], "Line 10: Financial Year")
	assert.Equal(suite.T(), "-150.00", results[9]["Receipt amount"], "Line 10: Receipt amount")
	assert.Equal(suite.T(), "-100.00", results[9]["Amount applied"], "Line 10: Amount applied")
	assert.Equal(suite.T(), "0.00", results[9]["Amount unapplied"], "Line 10: Amount unapplied")

	// new payment applied to correct client (with overpayment)
	assert.Equal(suite.T(), "Colette Correct", results[10]["Customer name"], "Line 11: Customer name")
	assert.Equal(suite.T(), courtRef3, results[10]["Customer number"], "Line 11: Customer number")
	assert.Equal(suite.T(), "3333", results[10]["SOP number"], "Line 11: SOP number")
	assert.Equal(suite.T(), "0470", results[10]["Entity"], "Line 11: Entity")
	assert.Equal(suite.T(), "99999999", results[10]["Receivables cost centre"], "Line 11: Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[10]["Receivables cost centre description"], "Line 11: Receivables cost centre description")
	assert.Equal(suite.T(), "1816102003", results[10]["Receivables account code"], "Line 11: Receivables account code")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES - SIRIUS SUPERVISION CONTROL ACCOUNT", results[10]["Account code description"], "Line 11: Account code description")
	assert.Equal(suite.T(), "OC"+inv5Ref, results[10]["Txn number"], "Line 11: Txn number")
	assert.Equal(suite.T(), "Online Card Payment", results[10]["Txn type"], "Line 11: Txn type")
	assert.Equal(suite.T(), yesterday.String(), results[10]["Receipt date"], "Line 11: Receipt date")
	assert.Equal(suite.T(), yesterday.String(), results[10]["Sirius upload date"], "Line 11: Sirius upload date")
	assert.Equal(suite.T(), yesterday.FinancialYear(), results[10]["Financial Year"], "Line 11: Financial Year")
	assert.Equal(suite.T(), "150.00", results[10]["Receipt amount"], "Line 11: Receipt amount")
	assert.Equal(suite.T(), "90.00", results[10]["Amount applied"], "Line 11: Amount applied")
	assert.Equal(suite.T(), "0.00", results[10]["Amount unapplied"], "Line 11: Amount unapplied")

	assert.Equal(suite.T(), "Colette Correct", results[11]["Customer name"], "Line 12: Customer name")
	assert.Equal(suite.T(), courtRef3, results[11]["Customer number"], "Line 12: Customer number")
	assert.Equal(suite.T(), "3333", results[11]["SOP number"], "Line 12: SOP number")
	assert.Equal(suite.T(), "0470", results[11]["Entity"], "Line 12: Entity")
	assert.Equal(suite.T(), "99999999", results[11]["Receivables cost centre"], "Line 12: Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[11]["Receivables cost centre description"], "Line 12: Receivables cost centre description")
	assert.Equal(suite.T(), "1816102005", results[11]["Receivables account code"], "Line 12: Receivables account code")
	assert.Equal(suite.T(), "CA - TRADE RECEIVABLES - ON ACCOUNT RECEIPTS – SIRIUS SUPERVISION", results[11]["Account code description"], "Line 12: Account code description")
	assert.Equal(suite.T(), "OC"+courtRef3, results[11]["Txn number"], "Line 12: Txn number")
	assert.Equal(suite.T(), "Online Card Payment", results[11]["Txn type"], "Line 12: Txn type")
	assert.Equal(suite.T(), yesterday.String(), results[11]["Receipt date"], "Line 12: Receipt date")
	assert.Equal(suite.T(), yesterday.String(), results[11]["Sirius upload date"], "Line 12: Sirius upload date")
	assert.Equal(suite.T(), yesterday.FinancialYear(), results[11]["Financial Year"], "Line 12: Financial Year")
	assert.Equal(suite.T(), "150.00", results[11]["Receipt amount"], "Line 12: Receipt amount")
	assert.Equal(suite.T(), "0.00", results[11]["Amount applied"], "Line 12: Amount applied")
	assert.Equal(suite.T(), "60.00", results[11]["Amount unapplied"], "Line 12: Amount unapplied")

	// refunds excess (not yet implemented)
}
