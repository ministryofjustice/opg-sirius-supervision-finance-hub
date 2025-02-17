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
	_, _ = suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, nil)

	//suite.seeder.CreatePayment(ctx, 100, yesterday.Date(), "12345678", shared.TransactionTypeOPGBACSPayment, yesterday.Date())
	suite.seeder.CreatePayment(ctx, 1500, yesterday.Date(), "12345678", shared.TransactionTypeMotoCardPayment, yesterday.Date())
	suite.seeder.CreatePayment(ctx, 2550, yesterday.Date(), "12345678", shared.TransactionTypeSupervisionBACSPayment, yesterday.Date())

	c := Client{suite.seeder.Conn}

	date := shared.NewDate(today.String())

	rows, err := c.Run(ctx, &ReceiptTransactions{
		Date: &date,
	})

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 5, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// OPG BACS Payments
	assert.Equal(suite.T(), "0470", results[0]["Entity"], "Entity - BACS Payments")
	assert.Equal(suite.T(), "99999999", results[0]["Cost Centre"], "Cost Centre - BACS Payments")
	assert.Equal(suite.T(), "1841102088", results[0]["Account"], "Account - BACS Payments")
	assert.Equal(suite.T(), "0000000", results[0]["Objective"], "Objective - BACS Payments")
	assert.Equal(suite.T(), "00000000", results[0]["Analysis"], "Analysis - BACS Payments")
	assert.Equal(suite.T(), "0000", results[0]["Intercompany"], "Intercompany - BACS Payments")
	assert.Equal(suite.T(), "000000", results[0]["Spare"], "Spare - BACS Payments")
	assert.Equal(suite.T(), "25.50", results[0]["Debit"], "Debit - BACS Payments")
	assert.Equal(suite.T(), "", results[0]["Credit"], "Credit - BACS Payments")
	assert.Equal(suite.T(), fmt.Sprintf("BACS Payment [%s]", yesterday.Date().Format("02/01/2006")), results[0]["Line description"], "Line description - BACS Payments")

	// OPG BACS Payments double
	assert.Equal(suite.T(), "0470", results[1]["Entity"], "Entity - BACS Payments 2")
	assert.Equal(suite.T(), "99999999", results[1]["Cost Centre"], "Cost Centre - BACS Payments 2")
	assert.Equal(suite.T(), "1816100000", results[1]["Account"], "Account - BACS Payments 2")
	assert.Equal(suite.T(), "0000000", results[1]["Objective"], "Objective - BACS Payments 2")
	assert.Equal(suite.T(), "00000000", results[1]["Analysis"], "Analysis - BACS Payments 2")
	assert.Equal(suite.T(), "0000", results[1]["Intercompany"], "Intercompany - BACS Payments 2")
	assert.Equal(suite.T(), "00000", results[1]["Spare"], "Spare - BACS Payments 2")
	assert.Equal(suite.T(), "", results[1]["Debit"], "Debit - BACS Payments 2")
	assert.Equal(suite.T(), "25.50", results[1]["Credit"], "Credit - BACS Payments 2")
	assert.Equal(suite.T(), fmt.Sprintf("BACS Payment [%s]", yesterday.Date().Format("02/01/2006")), results[1]["Line description"], "Line description - BACS Payments 2")

	// Moto Payments
	assert.Equal(suite.T(), "0470", results[2]["Entity"], "Entity - Moto Payments")
	assert.Equal(suite.T(), "99999999", results[2]["Cost Centre"], "Cost Centre - Moto Payments")
	assert.Equal(suite.T(), "1841102050", results[2]["Account"], "Account - Moto Payments")
	assert.Equal(suite.T(), "0000000", results[2]["Objective"], "Objective - Moto Payments")
	assert.Equal(suite.T(), "00000000", results[2]["Analysis"], "Analysis - Moto Payments")
	assert.Equal(suite.T(), "0000", results[2]["Intercompany"], "Intercompany - Moto Payments")
	assert.Equal(suite.T(), "000000", results[2]["Spare"], "Spare - Moto Payments")
	assert.Equal(suite.T(), "15.00", results[2]["Debit"], "Debit - Moto Payments")
	assert.Equal(suite.T(), "", results[2]["Credit"], "Credit - Moto Payments")
	assert.Equal(suite.T(), fmt.Sprintf("MOTO (Phone) Card Payment [%s]", yesterday.Date().Format("02/01/2006")), results[2]["Line description"], "Line description - Moto Payments")

	// Moto Payments -- reverse
	assert.Equal(suite.T(), "0470", results[3]["Entity"], "Entity - MOTO Payments 2")
	assert.Equal(suite.T(), "99999999", results[3]["Cost Centre"], "Cost Centre - MOTO Payments 2")
	assert.Equal(suite.T(), "1816100000", results[3]["Account"], "Account - MOTO Payments 2")
	assert.Equal(suite.T(), "0000000", results[3]["Objective"], "Objective - MOTO Payments 2")
	assert.Equal(suite.T(), "00000000", results[3]["Analysis"], "Analysis - MOTO Payments 2")
	assert.Equal(suite.T(), "0000", results[3]["Intercompany"], "Intercompany - MOTO Payments 2")
	assert.Equal(suite.T(), "00000", results[3]["Spare"], "Spare - MOTO Payments 2")
	assert.Equal(suite.T(), "", results[3]["Debit"], "Debit - MOTO Payments 2")
	assert.Equal(suite.T(), "15.00", results[3]["Credit"], "Credit - MOTO Payments 2")
	assert.Equal(suite.T(), fmt.Sprintf("MOTO (Phone) Card Payment [%s]", yesterday.Date().Format("02/01/2006")), results[3]["Line description"], "Line description - MOTO Payments 2")
}
