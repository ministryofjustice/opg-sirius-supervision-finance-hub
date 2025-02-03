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

	//suite.seeder.CreateInvoiceFeeRange(ctx, client1InvoiceId, "GENERAL")
	suite.seeder.CreatePayment(ctx, 100, yesterday.Date(), "12345678", shared.TransactionTypeOPGBACSPayment, yesterday.Date())

	c := Client{suite.seeder.Conn}

	date := shared.NewDate(today.String())

	rows, err := c.Run(ctx, &ReceiptTransactions{
		Date: &date,
	})

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// client 1 invoice 1
	assert.Equal(suite.T(), "0470", results[0]["Entity"], "Entity - client 1 invoice 1")
	assert.Equal(suite.T(), "10482009", results[0]["Cost Centre"], "Cost Centre - client 1 invoice 1")
	assert.Equal(suite.T(), "1816100000", results[0]["Account"], "Account - client 1 invoice 1")
	assert.Equal(suite.T(), "0000000", results[0]["Objective"], "Objective - client 1 invoice 1")
	assert.Equal(suite.T(), "00000000", results[0]["Analysis"], "Analysis - client 1 invoice 1")
	assert.Equal(suite.T(), "0000", results[0]["Intercompany"], "Intercompany - client 1 invoice 1")
	assert.Equal(suite.T(), "00000000", results[0]["Spare"], "Spare - client 1 invoice 1")
	assert.Equal(suite.T(), "", results[0]["Debit"], "Debit - client 1 invoice 1")
	assert.Equal(suite.T(), "100", results[0]["Credit"], "Credit - client 1 invoice 1")
	assert.Equal(suite.T(), fmt.Sprintf("BACS Payment [%s]", today.Date().Format("02/01/2006")), results[0]["Line description"], "Line description - client 1 invoice 1")
}
