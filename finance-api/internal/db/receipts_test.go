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
	twoYearsAgo := suite.seeder.Today().Sub(1, 0, 0)
	twoMonthsAgo := suite.seeder.Today().Sub(0, 2, 0)
	oneMonthAgo := suite.seeder.Today().Sub(0, 1, 0)

	// transaction timeline:
	// 1st invoice
	// paid in full
	// fee reduction unapplies 50%
	// 2nd invoice
	// 50% paid with reapply
	// 3rd invoice
	// payment covers 2nd and 3rd invoices, with excess
	clientID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12345678", "1234")
	_, inv1Ref := suite.seeder.CreateInvoice(ctx, clientID, shared.InvoiceTypeGA, nil, twoYearsAgo.StringPtr(), nil, nil, nil)
	// payment = 200
	suite.seeder.CreateFeeReduction(ctx, clientID, shared.FeeReductionTypeRemission, strconv.Itoa(twoYearsAgo.Date().Year()), 1, "A reduction")

	_, inv2Ref := suite.seeder.CreateInvoice(ctx, clientID, shared.InvoiceTypeS2, valToPtr("316.24"), twoMonthsAgo.StringPtr(), nil, nil, nil)
	_, inv3Ref := suite.seeder.CreateInvoice(ctx, clientID, shared.InvoiceTypeSO, valToPtr("70.00"), twoMonthsAgo.StringPtr(), nil, nil, nil)
	// payment = 300

	// excluded as out of range

	c := Client{suite.seeder.Conn}

	from := shared.NewDate(fourYearsAgo.String())
	to := shared.NewDate(yesterday.String())

	rows, err := c.Run(ctx, &Receipts{
		FromDate: &from,
		ToDate:   &to,
	})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 4, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// 1st invoice paid in full
	assert.Equal(suite.T(), "Ian Test", results[0]["Customer name"], "Customer name")
	assert.Equal(suite.T(), "12345678", results[0]["Customer number"], "Customer number")
	assert.Equal(suite.T(), "1234", results[0]["SOP number"], "SOP number")
	assert.Equal(suite.T(), "0470", results[0]["Entity"], "Entity")
	assert.Equal(suite.T(), "99999999", results[0]["Receivables cost centre"], "Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[0]["Receivables cost centre description"], "Receivables cost centre description")
	assert.Equal(suite.T(), "1816100000", results[0]["Receivables account code"], "Receivables account code")
	assert.Equal(suite.T(), "Invoice adjustments on receivables", results[0]["Account code description"], "Account code description")
	assert.Equal(suite.T(), "BC"+inv1Ref, results[0]["Txn number"], "Txn number")
	assert.Equal(suite.T(), "Bank Transfers", results[0]["Txn type"], "Txn type")
	assert.Equal(suite.T(), "15/03/2024", results[0]["Receipt date"], "Receipt date")
	assert.Equal(suite.T(), "15/03/2024", results[0]["Sirius upload date"], "Sirius upload date")
	assert.Equal(suite.T(), "2024", results[0]["Financial Year"], "Financial Year")
	assert.Equal(suite.T(), "200.00", results[0]["Receipt amount"], "Receipt amount")
	assert.Equal(suite.T(), "200.00", results[0]["Amount applied"], "Amount applied")
	assert.Equal(suite.T(), "0.00", results[0]["Amount unapplied"], "Amount unapplied")

	// fee reduction unapplies 50%
	assert.Equal(suite.T(), "Ian Test", results[0]["Customer name"], "Customer name")
	assert.Equal(suite.T(), "12345678", results[0]["Customer number"], "Customer number")
	assert.Equal(suite.T(), "1234", results[0]["SOP number"], "SOP number")
	assert.Equal(suite.T(), "0470", results[0]["Entity"], "Entity")
	assert.Equal(suite.T(), "99999999", results[0]["Receivables cost centre"], "Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[0]["Receivables cost centre description"], "Receivables cost centre description")
	assert.Equal(suite.T(), "1816100001", results[0]["Receivables account code"], "Receivables account code")
	assert.Equal(suite.T(), "Adjustments to unapplied monies on receivables", results[0]["Account code description"], "Account code description")
	assert.Equal(suite.T(), "UA"+inv2Ref, results[0]["Txn number"], "Txn number")
	assert.Equal(suite.T(), "Unapply", results[0]["Txn type"], "Txn type")
	assert.Equal(suite.T(), "", results[0]["Receipt date"], "Receipt date")
	assert.Equal(suite.T(), twoMonthsAgo.String(), results[0]["Sirius upload date"], "Sirius upload date")
	assert.Equal(suite.T(), twoMonthsAgo.FinancialYear(), results[0]["Financial Year"], "Financial Year")
	assert.Equal(suite.T(), "0.00", results[0]["Receipt amount"], "Receipt amount")
	assert.Equal(suite.T(), "0.00", results[0]["Amount applied"], "Amount applied")
	assert.Equal(suite.T(), "100.00", results[0]["Amount unapplied"], "Amount unapplied")

	// 2nd invoice 50% paid with reapply
	assert.Equal(suite.T(), "Ian Test", results[0]["Customer name"], "Customer name")
	assert.Equal(suite.T(), "12345678", results[0]["Customer number"], "Customer number")
	assert.Equal(suite.T(), "1234", results[0]["SOP number"], "SOP number")
	assert.Equal(suite.T(), "0470", results[0]["Entity"], "Entity")
	assert.Equal(suite.T(), "99999999", results[0]["Receivables cost centre"], "Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[0]["Receivables cost centre description"], "Receivables cost centre description")
	assert.Equal(suite.T(), "1816100000", results[0]["Receivables account code"], "Receivables account code")
	assert.Equal(suite.T(), "Invoice adjustments on receivables", results[0]["Account code description"], "Account code description")
	assert.Equal(suite.T(), "RA"+invoice2Ref, results[0]["Txn number"], "Txn number")
	assert.Equal(suite.T(), "Reapply", results[0]["Txn type"], "Txn type")
	assert.Equal(suite.T(), "15/03/2024", results[0]["Receipt date"], "Receipt date")
	assert.Equal(suite.T(), "15/03/2024", results[0]["Sirius upload date"], "Sirius upload date")
	assert.Equal(suite.T(), "2024", results[0]["Financial Year"], "Financial Year")
	assert.Equal(suite.T(), "0.00", results[0]["Receipt amount"], "Receipt amount")
	assert.Equal(suite.T(), "160.00", results[0]["Amount applied"], "Amount applied")
	assert.Equal(suite.T(), "0.00", results[0]["Amount unapplied"], "Amount unapplied")

	// payment covers 2nd invoice...
	assert.Equal(suite.T(), "Ian Test", results[0]["Customer name"], "Customer name")
	assert.Equal(suite.T(), "12345678", results[0]["Customer number"], "Customer number")
	assert.Equal(suite.T(), "1234", results[0]["SOP number"], "SOP number")
	assert.Equal(suite.T(), "0470", results[0]["Entity"], "Entity")
	assert.Equal(suite.T(), "99999999", results[0]["Receivables cost centre"], "Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[0]["Receivables cost centre description"], "Receivables cost centre description")
	assert.Equal(suite.T(), "1816100000", results[0]["Receivables account code"], "Receivables account code")
	assert.Equal(suite.T(), "Invoice adjustments on receivables", results[0]["Account code description"], "Account code description")
	assert.Equal(suite.T(), "PC"+invoice2Ref, results[0]["Txn number"], "Txn number")
	assert.Equal(suite.T(), "Moto payment", results[0]["Txn type"], "Txn type")
	assert.Equal(suite.T(), "15/03/2024", results[0]["Receipt date"], "Receipt date")
	assert.Equal(suite.T(), "15/03/2024", results[0]["Sirius upload date"], "Sirius upload date")
	assert.Equal(suite.T(), "2024", results[0]["Financial Year"], "Financial Year")
	assert.Equal(suite.T(), "300.00", results[0]["Receipt amount"], "Receipt amount")
	assert.Equal(suite.T(), "160.00", results[0]["Amount applied"], "Amount applied")
	assert.Equal(suite.T(), "0.00", results[0]["Amount unapplied"], "Amount unapplied")

	// ... and 3rd invoice
	assert.Equal(suite.T(), "Ian Test", results[0]["Customer name"], "Customer name")
	assert.Equal(suite.T(), "12345678", results[0]["Customer number"], "Customer number")
	assert.Equal(suite.T(), "1234", results[0]["SOP number"], "SOP number")
	assert.Equal(suite.T(), "0470", results[0]["Entity"], "Entity")
	assert.Equal(suite.T(), "99999999", results[0]["Receivables cost centre"], "Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[0]["Receivables cost centre description"], "Receivables cost centre description")
	assert.Equal(suite.T(), "1816100000", results[0]["Receivables account code"], "Receivables account code")
	assert.Equal(suite.T(), "Invoice adjustments on receivables", results[0]["Account code description"], "Account code description")
	assert.Equal(suite.T(), "PC"+invoice3Ref, results[0]["Txn number"], "Txn number")
	assert.Equal(suite.T(), "Moto payment", results[0]["Txn type"], "Txn type")
	assert.Equal(suite.T(), "15/03/2024", results[0]["Receipt date"], "Receipt date")
	assert.Equal(suite.T(), "15/03/2024", results[0]["Sirius upload date"], "Sirius upload date")
	assert.Equal(suite.T(), "2024", results[0]["Financial Year"], "Financial Year")
	assert.Equal(suite.T(), "300.00", results[0]["Receipt amount"], "Receipt amount")
	assert.Equal(suite.T(), "100.00", results[0]["Amount applied"], "Amount applied")
	assert.Equal(suite.T(), "0.00", results[0]["Amount unapplied"], "Amount unapplied")

	// ... and overpays excess
	assert.Equal(suite.T(), "Ian Test", results[0]["Customer name"], "Customer name")
	assert.Equal(suite.T(), "12345678", results[0]["Customer number"], "Customer number")
	assert.Equal(suite.T(), "1234", results[0]["SOP number"], "SOP number")
	assert.Equal(suite.T(), "0470", results[0]["Entity"], "Entity")
	assert.Equal(suite.T(), "99999999", results[0]["Receivables cost centre"], "Receivables cost centre")
	assert.Equal(suite.T(), "BALANCE SHEET", results[0]["Receivables cost centre description"], "Receivables cost centre description")
	assert.Equal(suite.T(), "1816100002", results[0]["Receivables account code"], "Receivables account code")
	assert.Equal(suite.T(), "Adjustments to on account receipts (overpayments)", results[0]["Account code description"], "Account code description")
	assert.Equal(suite.T(), "PC12345678", results[0]["Txn number"], "Txn number")
	assert.Equal(suite.T(), "Moto payment", results[0]["Txn type"], "Txn type")
	assert.Equal(suite.T(), "15/03/2024", results[0]["Receipt date"], "Receipt date")
	assert.Equal(suite.T(), "15/03/2024", results[0]["Sirius upload date"], "Sirius upload date")
	assert.Equal(suite.T(), "2024", results[0]["Financial Year"], "Financial Year")
	assert.Equal(suite.T(), "300.00", results[0]["Receipt amount"], "Receipt amount")
	assert.Equal(suite.T(), "40.00", results[0]["Amount applied"], "Amount applied")
	assert.Equal(suite.T(), "0.00", results[0]["Amount unapplied"], "Amount unapplied")

	// refunds excess (not yet implemented)
}
