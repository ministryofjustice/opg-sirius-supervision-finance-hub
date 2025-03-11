package db

import (
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/testhelpers"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_non_receipt_transactions() {
	ctx := suite.ctx
	general := "GENERAL"
	minimal := "MINIMAL"
	s2Amount := "300"
	s3Amount := "320"

	today := suite.seeder.Today()
	yesterday := suite.seeder.Today().Sub(0, 0, 1)
	twoMonthsAgo := suite.seeder.Today().Sub(0, 2, 0)
	threeMonthsAgo := suite.seeder.Today().Sub(0, 3, 0)
	oneYearAgo := suite.seeder.Today().Sub(1, 0, 0)

	// one client with one AD invoice, one minimal S3 invoice and an exemption
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12345678", "1234")
	suite.seeder.CreateOrder(ctx, client1ID, "ACTIVE")

	invoice1Id, _ := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, yesterday.StringPtr())
	suite.seeder.AddFeeRanges(ctx, invoice1Id, []testhelpers.FeeRange{{SupervisionLevel: "AD", FromDate: oneYearAgo.Date(), ToDate: today.Date()}})
	invoice2Id, _ := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeS3, &s3Amount, threeMonthsAgo.StringPtr(), nil, nil, &minimal, yesterday.StringPtr())
	suite.seeder.AddFeeRanges(ctx, invoice2Id, []testhelpers.FeeRange{{SupervisionLevel: "MINIMAL", FromDate: oneYearAgo.Date(), ToDate: today.Date()}})

	suite.seeder.CreateFeeReduction(ctx, client1ID, shared.FeeReductionTypeExemption, "2024", 2, "Test", yesterday.Date())

	// one client with one AD invoice, one general S2 invoice and a remission
	client2ID := suite.seeder.CreateClient(ctx, "Barry", "Test", "87654321", "4321")
	suite.seeder.CreateOrder(ctx, client1ID, "ACTIVE")

	invoice3Id, _ := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, yesterday.StringPtr())
	suite.seeder.AddFeeRanges(ctx, invoice3Id, []testhelpers.FeeRange{{SupervisionLevel: "AD", FromDate: oneYearAgo.Date(), ToDate: today.Date()}})
	invoice4Id, _ := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS2, &s2Amount, threeMonthsAgo.StringPtr(), nil, nil, &general, yesterday.StringPtr())
	suite.seeder.AddFeeRanges(ctx, invoice4Id, []testhelpers.FeeRange{{SupervisionLevel: "GENERAL", FromDate: oneYearAgo.Date(), ToDate: today.Date()}})

	suite.seeder.CreateFeeReduction(ctx, client2ID, shared.FeeReductionTypeHardship, "2024", 4, "Test", yesterday.Date())

	c := Client{suite.seeder.Conn}

	date := shared.NewDate(yesterday.String())

	rows, err := c.Run(ctx, &NonReceiptTransactions{
		Date: &date,
	})

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 13, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	assert.Equal(suite.T(), "0470", results[0]["Entity"], "Entity - AD invoice Credit")
	assert.Equal(suite.T(), "10482009", results[0]["Cost Centre"], "Cost Centre - AD invoice Credit")
	assert.Equal(suite.T(), "4481102093", results[0]["Account"], "Account - AD invoice Credit")
	assert.Equal(suite.T(), "0000000", results[0]["Objective"], "Objective - AD invoice Credit")
	assert.Equal(suite.T(), "00000000", results[0]["Analysis"], "Analysis - AD invoice Credit")
	assert.Equal(suite.T(), "0000", results[0]["Intercompany"], "Intercompany - AD invoice Credit")
	assert.Equal(suite.T(), "00000000", results[0]["Spare"], "Spare - AD invoice Credit")
	assert.Equal(suite.T(), "", results[0]["Debit"], "Debit - AD invoice Credit")
	assert.Equal(suite.T(), "200.00", results[0]["Credit"], "Credit - AD invoice Credit")
	assert.Equal(suite.T(), fmt.Sprintf("AD invoice [%s]", yesterday.Date().Format("02/01/2006")), results[0]["Line description"], "Line description - AD invoice Credit")

	assert.Equal(suite.T(), "0470", results[1]["Entity"], "Entity - AD invoice Debit")
	assert.Equal(suite.T(), "99999999", results[1]["Cost Centre"], "Cost Centre - AD invoice Debit")
	assert.Equal(suite.T(), "1816100000", results[1]["Account"], "Account - AD invoice Debit")
	assert.Equal(suite.T(), "0000000", results[1]["Objective"], "Objective - AD invoice Debit")
	assert.Equal(suite.T(), "00000000", results[1]["Analysis"], "Analysis - AD invoice Debit")
	assert.Equal(suite.T(), "0000", results[1]["Intercompany"], "Intercompany - AD invoice Debit")
	assert.Equal(suite.T(), "00000000", results[1]["Spare"], "Spare - AD invoice Debit")
	assert.Equal(suite.T(), "200.00", results[1]["Debit"], "Debit - AD invoice Debit")
	assert.Equal(suite.T(), "", results[1]["Credit"], "Credit - AD invoice Debit")
	assert.Equal(suite.T(), fmt.Sprintf("AD invoice [%s]", yesterday.Date().Format("02/01/2006")), results[1]["Line description"], "Line description - AD invoice Debit")

	assert.Equal(suite.T(), "0470", results[2]["Entity"], "Entity - S2 invoice Credit")
	assert.Equal(suite.T(), "10482009", results[2]["Cost Centre"], "Cost Centre - S2 invoice Credit")
	assert.Equal(suite.T(), "4481102094", results[2]["Account"], "Account - S2 invoice Credit")
	assert.Equal(suite.T(), "0000000", results[2]["Objective"], "Objective - S2 invoice Credit")
	assert.Equal(suite.T(), "00000000", results[2]["Analysis"], "Analysis - S2 invoice Credit")
	assert.Equal(suite.T(), "0000", results[2]["Intercompany"], "Intercompany - S2 invoice Credit")
	assert.Equal(suite.T(), "00000000", results[2]["Spare"], "Spare - S2 invoice Credit")
	assert.Equal(suite.T(), "", results[2]["Debit"], "Debit - S2 invoice Credit")
	assert.Equal(suite.T(), "300.00", results[2]["Credit"], "Credit - S2 invoice Credit")
	assert.Equal(suite.T(), fmt.Sprintf("S2 invoice [%s]", yesterday.Date().Format("02/01/2006")), results[2]["Line description"], "Line description - S2 invoice Credit")

	assert.Equal(suite.T(), "0470", results[3]["Entity"], "Entity - S2 invoice Debit")
	assert.Equal(suite.T(), "99999999", results[3]["Cost Centre"], "Cost Centre - S2 invoice Debit")
	assert.Equal(suite.T(), "1816100000", results[3]["Account"], "Account - S2 invoice Debit")
	assert.Equal(suite.T(), "0000000", results[3]["Objective"], "Objective - S2 invoice Debit")
	assert.Equal(suite.T(), "00000000", results[3]["Analysis"], "Analysis - S2 invoice Debit")
	assert.Equal(suite.T(), "0000", results[3]["Intercompany"], "Intercompany - S2 invoice Debit")
	assert.Equal(suite.T(), "00000000", results[3]["Spare"], "Spare - S2 invoice Debit")
	assert.Equal(suite.T(), "300.00", results[3]["Debit"], "Debit - S2 invoice Debit")
	assert.Equal(suite.T(), "", results[3]["Credit"], "Credit - S2 invoice Debit")
	assert.Equal(suite.T(), fmt.Sprintf("S2 invoice [%s]", yesterday.Date().Format("02/01/2006")), results[3]["Line description"], "Line description - S2 invoice Debit")

	assert.Equal(suite.T(), "0470", results[4]["Entity"], "Entity - S3 invoice Credit")
	assert.Equal(suite.T(), "10482009", results[4]["Cost Centre"], "Cost Centre - S3 invoice Credit")
	assert.Equal(suite.T(), "4481102099", results[4]["Account"], "Account - S3 invoice Credit")
	assert.Equal(suite.T(), "0000000", results[4]["Objective"], "Objective - S3 invoice Credit")
	assert.Equal(suite.T(), "00000000", results[4]["Analysis"], "Analysis - S3 invoice Credit")
	assert.Equal(suite.T(), "0000", results[4]["Intercompany"], "Intercompany - S3 invoice Credit")
	assert.Equal(suite.T(), "00000000", results[4]["Spare"], "Spare - S3 invoice Credit")
	assert.Equal(suite.T(), "", results[4]["Debit"], "Debit - S3 invoice Credit")
	assert.Equal(suite.T(), "320.00", results[4]["Credit"], "Credit - S3 invoice Credit")
	assert.Equal(suite.T(), fmt.Sprintf("S3 invoice [%s]", yesterday.Date().Format("02/01/2006")), results[4]["Line description"], "Line description - S3 invoice Credit")

	assert.Equal(suite.T(), "0470", results[5]["Entity"], "Entity - S3 invoice Debit")
	assert.Equal(suite.T(), "99999999", results[5]["Cost Centre"], "Cost Centre - S3 invoice Debit")
	assert.Equal(suite.T(), "1816100000", results[5]["Account"], "Account - S3 invoice Debit")
	assert.Equal(suite.T(), "0000000", results[5]["Objective"], "Objective - S3 invoice Debit")
	assert.Equal(suite.T(), "00000000", results[5]["Analysis"], "Analysis - S3 invoice Debit")
	assert.Equal(suite.T(), "0000", results[5]["Intercompany"], "Intercompany - S3 invoice Debit")
	assert.Equal(suite.T(), "00000000", results[5]["Spare"], "Spare - S3 invoice Debit")
	assert.Equal(suite.T(), "320.00", results[5]["Debit"], "Debit - S3 invoice Debit")
	assert.Equal(suite.T(), "", results[5]["Credit"], "Credit - S3 invoice Debit")
	assert.Equal(suite.T(), fmt.Sprintf("S3 invoice [%s]", yesterday.Date().Format("02/01/2006")), results[5]["Line description"], "Line description - S3 invoice Debit")

	assert.Equal(suite.T(), "0470", results[6]["Entity"], "Entity - AD Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "10482009", results[6]["Cost Centre"], "Cost Centre - AD Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "4481102114", results[6]["Account"], "Account - AD Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "0000000", results[6]["Objective"], "Objective - AD Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "00000000", results[6]["Analysis"], "Analysis - AD Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "0000", results[6]["Intercompany"], "Intercompany - AD Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "00000000", results[6]["Spare"], "Spare - AD Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "100.00", results[6]["Debit"], "Debit - AD Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "", results[6]["Credit"], "Credit - AD Remissions & Exemptions Debit")
	assert.Equal(suite.T(), fmt.Sprintf("AD Rem/Exem [%s]", yesterday.Date().Format("02/01/2006")), results[6]["Line description"], "Line description - AD Remissions & Exemptions Debit")

	assert.Equal(suite.T(), "0470", results[7]["Entity"], "Entity - AD Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "99999999", results[7]["Cost Centre"], "Cost Centre - AD Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "1816100000", results[7]["Account"], "Account - AD Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "0000000", results[7]["Objective"], "Objective - AD Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "00000000", results[7]["Analysis"], "Analysis - AD Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "0000", results[7]["Intercompany"], "Intercompany - AD Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "00000000", results[7]["Spare"], "Spare - AD Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "", results[7]["Debit"], "Debit - AD Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "100.00", results[7]["Credit"], "Credit - AD Remissions & Exemptions Credit")
	assert.Equal(suite.T(), fmt.Sprintf("AD Rem/Exem [%s]", yesterday.Date().Format("02/01/2006")), results[7]["Line description"], "Line description - AD Remissions & Exemptions Credit")

	assert.Equal(suite.T(), "0470", results[8]["Entity"], "Entity - General Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "10482009", results[8]["Cost Centre"], "Cost Centre - General Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "4481102115", results[8]["Account"], "Account - General Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "0000000", results[8]["Objective"], "Objective - General Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "00000000", results[8]["Analysis"], "Analysis - General Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "0000", results[8]["Intercompany"], "Intercompany - General Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "00000000", results[8]["Spare"], "Spare - General Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "300.00", results[8]["Debit"], "Debit - General Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "", results[8]["Credit"], "Credit - General Remissions & Exemptions Debit")
	assert.Equal(suite.T(), fmt.Sprintf("Gen Rem/Exem [%s]", yesterday.Date().Format("02/01/2006")), results[8]["Line description"], "Line description - General Remissions & Exemptions Debit")

	assert.Equal(suite.T(), "0470", results[9]["Entity"], "Entity - General Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "99999999", results[9]["Cost Centre"], "Cost Centre - General Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "1816100000", results[9]["Account"], "Account - General Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "0000000", results[9]["Objective"], "Objective - General Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "00000000", results[9]["Analysis"], "Analysis - General Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "0000", results[9]["Intercompany"], "Intercompany - General Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "00000000", results[9]["Spare"], "Spare - General Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "", results[9]["Debit"], "Debit - General Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "300.00", results[9]["Credit"], "Credit - General Remissions & Exemptions Credit")
	assert.Equal(suite.T(), fmt.Sprintf("Gen Rem/Exem [%s]", yesterday.Date().Format("02/01/2006")), results[9]["Line description"], "Line description - General Remissions & Exemptions Credit")

	assert.Equal(suite.T(), "0470", results[10]["Entity"], "Entity - Minimal Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "10482009", results[10]["Cost Centre"], "Cost Centre - Minimal Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "4481102120", results[10]["Account"], "Account - Minimal Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "0000000", results[10]["Objective"], "Objective - Minimal Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "00000000", results[10]["Analysis"], "Analysis - Minimal Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "0000", results[10]["Intercompany"], "Intercompany - Minimal Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "00000000", results[10]["Spare"], "Spare - Minimal Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "320.00", results[10]["Debit"], "Debit - Minimal Remissions & Exemptions Debit")
	assert.Equal(suite.T(), "", results[10]["Credit"], "Credit - Minimal Remissions & Exemptions Debit")
	assert.Equal(suite.T(), fmt.Sprintf("Min Rem/Exem [%s]", yesterday.Date().Format("02/01/2006")), results[10]["Line description"], "Line description - Minimal Remissions & Exemptions Debit")

	assert.Equal(suite.T(), "0470", results[11]["Entity"], "Entity - Minimal Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "99999999", results[11]["Cost Centre"], "Cost Centre - Minimal Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "1816100000", results[11]["Account"], "Account - Minimal Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "0000000", results[11]["Objective"], "Objective - Minimal Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "00000000", results[11]["Analysis"], "Analysis - Minimal Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "0000", results[11]["Intercompany"], "Intercompany - Minimal Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "00000000", results[11]["Spare"], "Spare - Minimal Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "", results[11]["Debit"], "Debit - Minimal Remissions & Exemptions Credit")
	assert.Equal(suite.T(), "320.00", results[11]["Credit"], "Credit - Minimal Remissions & Exemptions Credit")
	assert.Equal(suite.T(), fmt.Sprintf("Min Rem/Exem [%s]", yesterday.Date().Format("02/01/2006")), results[11]["Line description"], "Line description - Minimal Remissions & Exemptions Credit")
}
