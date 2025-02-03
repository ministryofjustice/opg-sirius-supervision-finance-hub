package db

import (
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_receipt_transactions() {
	ctx := suite.ctx

	today := suite.seeder.Today()
	yesterday := suite.seeder.Today().Sub(0, 0, 1)
	twoMonthsAgo := suite.seeder.Today().Sub(0, 2, 0)

	// one client with one invoice and a BACS payment - credit
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12345678", "1234")
	suite.seeder.CreateOrder(ctx, client1ID, "ACTIVE")
	_, _ = suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil)

	suite.seeder.CreatePayment(ctx, 100, yesterday.Date(), "12345678", shared.TransactionTypeOPGBACSPayment, yesterday.Date())
	suite.seeder.CreatePayment(ctx, 100, yesterday.Date(), "12345678", shared.TransactionTypeMotoCardPayment, yesterday.Date())

	c := Client{suite.seeder.Conn}

	date := shared.NewDate(today.String())

	rows, err := c.Run(ctx, &ReceiptTransactions{
		Date: &date,
	})

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// OPG BACS Payments
	assert.Equal(suite.T(), "0470", results[0]["Entity"], "Entity - BACS Payments")
	assert.Equal(suite.T(), "10482009", results[0]["Cost Centre"], "Cost Centre - BACS Payments")
	assert.Equal(suite.T(), "1816100000", results[0]["Account"], "Account - BACS Payments")
	assert.Equal(suite.T(), "0000000", results[0]["Objective"], "Objective - BACS Payments")
	assert.Equal(suite.T(), "00000000", results[0]["Analysis"], "Analysis - BACS Payments")
	assert.Equal(suite.T(), "0000", results[0]["Intercompany"], "Intercompany - BACS Payments")
	assert.Equal(suite.T(), "00000000", results[0]["Spare"], "Spare - BACS Payments")
	assert.Equal(suite.T(), "", results[0]["Debit"], "Debit - BACS Payments")
	assert.Equal(suite.T(), "100", results[0]["Credit"], "Credit - BACS Payments")
	assert.Equal(suite.T(), fmt.Sprintf("BACS Payment [%s]", today.Date().Format("02/01/2006")), results[0]["Line description"], "Line description - BACS Payments")

	// Moto Payments
	assert.Equal(suite.T(), "0470", results[1]["Entity"], "Entity - Moto Payments")
	assert.Equal(suite.T(), "99999999", results[1]["Cost Centre"], "Cost Centre - Moto Payments")
	assert.Equal(suite.T(), "1816100000", results[1]["Account"], "Account - Moto Payments")
	assert.Equal(suite.T(), "0000000", results[1]["Objective"], "Objective - Moto Payments")
	assert.Equal(suite.T(), "00000000", results[1]["Analysis"], "Analysis - Moto Payments")
	assert.Equal(suite.T(), "0000", results[1]["Intercompany"], "Intercompany - Moto Payments")
	assert.Equal(suite.T(), "00000000", results[1]["Spare"], "Spare - Moto Payments")
	assert.Equal(suite.T(), "100", results[1]["Debit"], "Debit - Moto Payments")
	assert.Equal(suite.T(), "", results[1]["Credit"], "Credit - Moto Payments")
	assert.Equal(suite.T(), fmt.Sprintf("MOTO (Phone) Card Payment [%s]", today.Date().Format("02/01/2006")), results[1]["Line description"], "Line description - Moto Payments")
}
