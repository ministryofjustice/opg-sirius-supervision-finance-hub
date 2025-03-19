package db

import (
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"strconv"
)

func (suite *IntegrationSuite) Test_receipt_transactions() {
	ctx := suite.ctx

	today := suite.seeder.Today()
	yesterday := suite.seeder.Today().Sub(0, 0, 1)
	twoMonthsAgo := suite.seeder.Today().Sub(0, 2, 0)

	// one client with an invoice with a MOTO card payment, supervision BACS payment and an AD invoice with an exemption. This also creates an unapply
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12345678", "1234")
	suite.seeder.CreateOrder(ctx, client1ID, "ACTIVE")
	_, _ = suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, nil)

	suite.seeder.CreatePayment(ctx, 1500, yesterday.Date(), "12345678", shared.TransactionTypeMotoCardPayment, yesterday.Date())
	suite.seeder.CreatePayment(ctx, 2550, yesterday.Date(), "12345678", shared.TransactionTypeSupervisionBACSPayment, yesterday.Date())

	suite.seeder.CreateFeeReduction(ctx, client1ID, shared.FeeReductionTypeExemption, strconv.Itoa(yesterday.Date().Year()-1), 4, "", yesterday.Date())

	// one client with an invoice with a MOTO card payment, an OPG BACS payment and an S2 invoice with an approved credit memo
	client2ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "87654321", "4321")
	suite.seeder.CreateOrder(ctx, client2ID, "ACTIVE")
	invoice2ID, _ := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS2, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, nil)

	suite.seeder.CreatePayment(ctx, 120, yesterday.Date(), "87654321", shared.TransactionTypeOPGBACSPayment, yesterday.Date())
	suite.seeder.CreatePayment(ctx, 1500, yesterday.Date(), "87654321", shared.TransactionTypeMotoCardPayment, yesterday.Date())

	suite.seeder.CreateAdjustment(ctx, client2ID, invoice2ID, shared.AdjustmentTypeCreditMemo, -2500, "", yesterday.DatePtr())

	// one client with GA invoice, direct debit payment, online card payment and reapply
	client3ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12348765", "2314")
	suite.seeder.CreateOrder(ctx, client2ID, "ACTIVE")
	_, _ = suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeGA, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, nil)

	suite.seeder.CreatePayment(ctx, 4020, yesterday.Date(), "12348765", shared.TransactionTypeDirectDebitPayment, yesterday.Date())
	suite.seeder.CreatePayment(ctx, 1700, yesterday.Date(), "12348765", shared.TransactionTypeOnlineCardPayment, yesterday.Date())
	suite.seeder.CreatePayment(ctx, 1500, yesterday.Date(), "12348765", shared.TransactionTypeReapply, yesterday.Date())

	c := Client{suite.seeder.Conn}

	date := shared.NewDate(today.String())

	rows, err := c.Run(ctx, &ReceiptTransactions{
		Date: &date,
	})

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 11, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	assert.Equal(suite.T(), "=\"0470\"", results[0]["Entity"], "Entity - MOTO card Payments Debit")
	assert.Equal(suite.T(), "99999999", results[0]["Cost Centre"], "Cost Centre - MOTO card Payments Debit")
	assert.Equal(suite.T(), "1841102050", results[0]["Account"], "Account - MOTO card Payments Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[0]["Objective"], "Objective - MOTO card Payments Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[0]["Analysis"], "Analysis - MOTO card Payments Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[0]["Intercompany"], "Intercompany - MOTO card Payments Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[0]["Spare"], "Spare - MOTO card Payments Debit")
	assert.Equal(suite.T(), "30.00", results[0]["Debit"], "Debit - MOTO card Payments Debit")
	assert.Equal(suite.T(), "", results[0]["Credit"], "Credit - MOTO card Payments Debit")
	assert.Equal(suite.T(), fmt.Sprintf("MOTO card [%s]", yesterday.Date().Format("02/01/2006")), results[0]["Line description"], "Line description - MOTO card Payments Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[1]["Entity"], "Entity - MOTO card Payments Credit")
	assert.Equal(suite.T(), "99999999", results[1]["Cost Centre"], "Cost Centre - MOTO card Payments Credit")
	assert.Equal(suite.T(), "1816100000", results[1]["Account"], "Account - MOTO card Payments Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[1]["Objective"], "Objective - MOTO card Payments Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[1]["Analysis"], "Analysis - MOTO card Payments Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[1]["Intercompany"], "Intercompany - MOTO card Payments Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[1]["Spare"], "Spare - MOTO card Payments Credit")
	assert.Equal(suite.T(), "", results[1]["Debit"], "Debit - MOTO card Payments Credit")
	assert.Equal(suite.T(), "30.00", results[1]["Credit"], "Credit - MOTO card Payments Credit")
	assert.Equal(suite.T(), fmt.Sprintf("MOTO card [%s]", yesterday.Date().Format("02/01/2006")), results[1]["Line description"], "Line description - MOTO card Payments Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[2]["Entity"], "Entity - Online card Payments Debit")
	assert.Equal(suite.T(), "99999999", results[2]["Cost Centre"], "Cost Centre - Online card Payments Debit")
	assert.Equal(suite.T(), "1841102050", results[2]["Account"], "Account - Online card Payments Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[2]["Objective"], "Objective - Online card Payments Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[2]["Analysis"], "Analysis - Online card Payments Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[2]["Intercompany"], "Intercompany - Online card Payments Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[2]["Spare"], "Spare - Online card Payments Debit")
	assert.Equal(suite.T(), "17.00", results[2]["Debit"], "Debit - Online card Payments Debit")
	assert.Equal(suite.T(), "", results[2]["Credit"], "Credit - Online card Payments Debit")
	assert.Equal(suite.T(), fmt.Sprintf("Online card [%s]", yesterday.Date().Format("02/01/2006")), results[2]["Line description"], "Line description - Online card Payments Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[3]["Entity"], "Entity - Online card Payments Credit")
	assert.Equal(suite.T(), "99999999", results[3]["Cost Centre"], "Cost Centre - Online card Payments Credit")
	assert.Equal(suite.T(), "1816100000", results[3]["Account"], "Account - Online card Payments Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[3]["Objective"], "Objective - Online card Payments Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[3]["Analysis"], "Analysis - Online card Payments Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[3]["Intercompany"], "Intercompany - Online card Payments Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[3]["Spare"], "Spare - Online card Payments Credit")
	assert.Equal(suite.T(), "", results[3]["Debit"], "Debit - Online card Payments Credit")
	assert.Equal(suite.T(), "17.00", results[3]["Credit"], "Credit - Online card Payments Credit")
	assert.Equal(suite.T(), fmt.Sprintf("Online card [%s]", yesterday.Date().Format("02/01/2006")), results[3]["Line description"], "Line description - Online card Payments Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[4]["Entity"], "Entity - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "99999999", results[4]["Cost Centre"], "Cost Centre - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "1841102050", results[4]["Account"], "Account - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[4]["Objective"], "Objective - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[4]["Analysis"], "Analysis - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[4]["Intercompany"], "Intercompany - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[4]["Spare"], "Spare - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "1.20", results[4]["Debit"], "Debit - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "", results[4]["Credit"], "Credit - OPG BACS Payments Debit")
	assert.Equal(suite.T(), fmt.Sprintf("OPG BACS [%s]", yesterday.Date().Format("02/01/2006")), results[4]["Line description"], "Line description - OPG BACS Payments Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[5]["Entity"], "Entity - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "99999999", results[5]["Cost Centre"], "Cost Centre - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "1816100000", results[5]["Account"], "Account - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[5]["Objective"], "Objective - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[5]["Analysis"], "Analysis - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[5]["Intercompany"], "Intercompany - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[5]["Spare"], "Spare - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "", results[5]["Debit"], "Debit - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "1.20", results[5]["Credit"], "Credit - OPG BACS Payments Credit")
	assert.Equal(suite.T(), fmt.Sprintf("OPG BACS [%s]", yesterday.Date().Format("02/01/2006")), results[5]["Line description"], "Line description - OPG BACS Payments Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[6]["Entity"], "Entity - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "99999999", results[6]["Cost Centre"], "Cost Centre - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "1841102088", results[6]["Account"], "Account - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[6]["Objective"], "Objective - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[6]["Analysis"], "Analysis - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[6]["Intercompany"], "Intercompany - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[6]["Spare"], "Spare - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "25.50", results[6]["Debit"], "Debit - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "", results[6]["Credit"], "Credit - OPG BACS Payments Debit")
	assert.Equal(suite.T(), fmt.Sprintf("Supervision BACS [%s]", yesterday.Date().Format("02/01/2006")), results[6]["Line description"], "Line description - OPG BACS Payments Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[7]["Entity"], "Entity - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "99999999", results[7]["Cost Centre"], "Cost Centre - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "1816100000", results[7]["Account"], "Account - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[7]["Objective"], "Objective - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[7]["Analysis"], "Analysis - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[7]["Intercompany"], "Intercompany - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[7]["Spare"], "Spare - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "", results[7]["Debit"], "Debit - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "25.50", results[7]["Credit"], "Credit - OPG BACS Payments Credit")
	assert.Equal(suite.T(), fmt.Sprintf("Supervision BACS [%s]", yesterday.Date().Format("02/01/2006")), results[7]["Line description"], "Line description - OPG BACS Payments Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[8]["Entity"], "Entity - Direct Debit Debit")
	assert.Equal(suite.T(), "99999999", results[8]["Cost Centre"], "Cost Centre - Direct Debit Debit")
	assert.Equal(suite.T(), "1841102050", results[8]["Account"], "Account - Direct Debit Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[8]["Objective"], "Objective - Direct Debit Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[8]["Analysis"], "Analysis - Direct Debit Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[8]["Intercompany"], "Intercompany - Direct Debit Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[8]["Spare"], "Spare - Direct Debit Debit")
	assert.Equal(suite.T(), "40.20", results[8]["Debit"], "Debit - Direct Debit Debit")
	assert.Equal(suite.T(), "", results[8]["Credit"], "Credit - Direct Debit Debit")
	assert.Equal(suite.T(), fmt.Sprintf("Direct debit [%s]", yesterday.Date().Format("02/01/2006")), results[8]["Line description"], "Line description - Direct Debit Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[9]["Entity"], "Entity - Direct Debit Credit")
	assert.Equal(suite.T(), "99999999", results[9]["Cost Centre"], "Cost Centre - Direct Debit Credit")
	assert.Equal(suite.T(), "1816100000", results[9]["Account"], "Account - Direct Debit Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[9]["Objective"], "Objective - Direct Debit Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[9]["Analysis"], "Analysis - Direct Debit Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[9]["Intercompany"], "Intercompany - Direct Debit Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[9]["Spare"], "Spare - Direct Debit Credit")
	assert.Equal(suite.T(), "", results[9]["Debit"], "Debit - Direct Debit Credit")
	assert.Equal(suite.T(), "40.20", results[9]["Credit"], "Credit - Direct Debit Credit")
	assert.Equal(suite.T(), fmt.Sprintf("Direct debit [%s]", yesterday.Date().Format("02/01/2006")), results[9]["Line description"], "Line description - Direct Debit Credit")

}
