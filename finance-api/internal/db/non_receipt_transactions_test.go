package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_non_receipt_transactions() {
	ctx := suite.ctx

	//today := suite.seeder.Today()
	yesterday := suite.seeder.Today().Sub(0, 0, 1)
	twoMonthsAgo := suite.seeder.Today().Sub(0, 2, 0)

	// one client with one invoice and an exemption
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12345678", "1234")
	suite.seeder.CreateOrder(ctx, client1ID, "ACTIVE")
	client1InvoiceId, _ := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil)

	suite.seeder.CreateInvoiceFeeRange(ctx, client1InvoiceId, "AD")
	suite.seeder.CreateFeeReduction(ctx, client1ID, shared.FeeReductionTypeExemption, "2022", 4, "Test", yesterday.Date())

	c := Client{suite.seeder.Conn}

	date := shared.NewDate(yesterday.String())

	rows, err := c.Run(ctx, &NonReceiptTransactions{
		Date: &date,
	})

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// client 1 invoice 1
	assert.Equal(suite.T(), "0470", results[0]["Entity"], "Entity - client 1 invoice 1")
	assert.Equal(suite.T(), "10482009", results[0]["Cost Centre"], "Cost Centre - client 1 invoice 1")
	assert.Equal(suite.T(), "4481102114", results[0]["Account"], "Account - client 1 invoice 1")
	assert.Equal(suite.T(), "0000000", results[0]["Objective"], "Objective - client 1 invoice 1")
	assert.Equal(suite.T(), "00000000", results[0]["Analysis"], "Analysis - client 1 invoice 1")
	assert.Equal(suite.T(), "0000", results[0]["Intercompany"], "Intercompany - client 1 invoice 1")
	assert.Equal(suite.T(), "00000000", results[0]["Spare"], "Spare - client 1 invoice 1")
	assert.Equal(suite.T(), "", results[0]["Debit"], "Debit - client 1 invoice 1")
	assert.Equal(suite.T(), "", results[0]["Credit"], "Credit - client 1 invoice 1")
	assert.Equal(suite.T(), "10000", results[0]["Line description"], "Line description - client 1 invoice 1")
}
