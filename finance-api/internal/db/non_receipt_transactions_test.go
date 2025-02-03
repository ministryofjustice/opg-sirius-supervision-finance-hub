package db

import (
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_non_receipt_transactions() {
	ctx := suite.ctx
	general := "GENERAL"
	amount := "320"

	yesterday := suite.seeder.Today().Sub(0, 0, 1)
	twoMonthsAgo := suite.seeder.Today().Sub(0, 2, 0)
	threeMonthsAgo := suite.seeder.Today().Sub(0, 3, 0)

	// one client with three invoices of different types and an exemption
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12345678", "1234")
	suite.seeder.CreateOrder(ctx, client1ID, "ACTIVE")
	invoice1Id, _ := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil)
	suite.seeder.CreateInvoiceFeeRange(ctx, invoice1Id, "AD")

	invoice2Id, _ := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeS2, &amount, threeMonthsAgo.StringPtr(), nil, nil, &general)
	suite.seeder.CreateInvoiceFeeRange(ctx, invoice2Id, "GENERAL")

	invoice3Id, _ := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil)
	suite.seeder.CreateInvoiceFeeRange(ctx, invoice3Id, "AD")

	suite.seeder.CreateFeeReduction(ctx, client1ID, shared.FeeReductionTypeExemption, "2022", 4, "Test", yesterday.Date())

	c := Client{suite.seeder.Conn}

	date := shared.NewDate(yesterday.String())

	rows, err := c.Run(ctx, &NonReceiptTransactions{
		Date: &date,
	})

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// AD Remissions & Exemptions
	assert.Equal(suite.T(), "0470", results[0]["Entity"], "Entity - AD Remissions & Exemptions")
	assert.Equal(suite.T(), "10482009", results[0]["Cost Centre"], "Cost Centre - AD Remissions & Exemptions")
	assert.Equal(suite.T(), "4481102114", results[0]["Account"], "Account - AD Remissions & Exemptions")
	assert.Equal(suite.T(), "0000000", results[0]["Objective"], "Objective - AD Remissions & Exemptions")
	assert.Equal(suite.T(), "00000000", results[0]["Analysis"], "Analysis - AD Remissions & Exemptions")
	assert.Equal(suite.T(), "0000", results[0]["Intercompany"], "Intercompany - AD Remissions & Exemptions")
	assert.Equal(suite.T(), "00000000", results[0]["Spare"], "Spare - AD Remissions & Exemptions")
	assert.Equal(suite.T(), "", results[0]["Debit"], "Debit - AD Remissions & Exemptions")
	assert.Equal(suite.T(), "20000", results[0]["Credit"], "Credit - AD Remissions & Exemptions")
	assert.Equal(suite.T(), fmt.Sprintf("AD Rem/Exem [%s]", yesterday.Date().Format("02/01/2006")), results[0]["Line description"], "Line description - AD Remissions & Exemptions")

	assert.Equal(suite.T(), "0470", results[1]["Entity"], "Entity - S2 Remissions & Exemptions")
	assert.Equal(suite.T(), "10482009", results[1]["Cost Centre"], "Cost Centre - S2 Remissions & Exemptions")
	assert.Equal(suite.T(), "4481102115", results[1]["Account"], "Account - S2 Remissions & Exemptions")
	assert.Equal(suite.T(), "0000000", results[1]["Objective"], "Objective - S2 Remissions & Exemptions")
	assert.Equal(suite.T(), "00000000", results[1]["Analysis"], "Analysis - S2 Remissions & Exemptions")
	assert.Equal(suite.T(), "0000", results[1]["Intercompany"], "Intercompany - S2 Remissions & Exemptions")
	assert.Equal(suite.T(), "00000000", results[1]["Spare"], "Spare - S2 Remissions & Exemptions")
	assert.Equal(suite.T(), "", results[1]["Debit"], "Debit - S2 Remissions & Exemptions")
	assert.Equal(suite.T(), "32000", results[1]["Credit"], "Credit - S2 Remissions & Exemptions")
	assert.Equal(suite.T(), fmt.Sprintf("Gen Rem/Exem [%s]", yesterday.Date().Format("02/01/2006")), results[1]["Line description"], "Line description - S2 Remissions & Exemptions")
}
