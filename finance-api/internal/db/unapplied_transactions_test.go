package db

import (
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"strconv"
)

func (suite *IntegrationSuite) Test_unapplied_transactions() {
	ctx := suite.ctx

	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)
	twoMonthsAgo := today.Sub(0, 2, 0)
	threeYearsAgo := today.Sub(3, 0, 0)
	general := "320.00"
	minimal := "10.00"

	// 15.00 unapply
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12345678", "1234")
	suite.seeder.CreateOrder(ctx, client1ID, "ACTIVE")
	_, _ = suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, twoMonthsAgo.StringPtr())
	suite.seeder.CreatePayment(ctx, 1500, yesterday.Date(), "12345678", shared.TransactionTypeMotoCardPayment, yesterday.Date(), 0)
	_ = suite.seeder.CreateFeeReduction(ctx, client1ID, shared.FeeReductionTypeExemption, strconv.Itoa(yesterday.Date().Year()-1), 4, "", yesterday.Date())

	// an existing 100.00 unapply that has 10.00 reapplied
	client2ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "87654321", "4321")
	suite.seeder.CreateOrder(ctx, client2ID, "ACTIVE")
	_, _ = suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS2, &general, threeYearsAgo.StringPtr(), nil, nil, nil, threeYearsAgo.StringPtr())
	suite.seeder.CreatePayment(ctx, 42000, threeYearsAgo.Date(), "87654321", shared.TransactionTypeOPGBACSPayment, threeYearsAgo.Date(), 0)
	_ = suite.seeder.CreateFeeReduction(ctx, client2ID, shared.FeeReductionTypeExemption, strconv.Itoa(threeYearsAgo.Date().Year()-1), 2, "", threeYearsAgo.Date())
	_, _ = suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS3, &minimal, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())

	// 12.34 unapply and reapply on the same day
	client3ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "33333333", "4321")
	suite.seeder.CreateOrder(ctx, client3ID, "ACTIVE")
	invoiceID, _ := suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeS2, &general, twoMonthsAgo.StringPtr(), nil, nil, nil, twoMonthsAgo.StringPtr())
	suite.seeder.CreatePayment(ctx, 32000, yesterday.Date(), "33333333", shared.TransactionTypeMotoCardPayment, yesterday.Date(), 0)
	suite.seeder.CreateAdjustment(ctx, client3ID, invoiceID, shared.AdjustmentTypeCreditMemo, 1234, "Credit memo", yesterday.DatePtr())
	_, _ = suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeS2, &general, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())

	c := Client{suite.seeder.Conn}

	date := shared.NewDate(yesterday.String())

	rows, err := c.Run(ctx, NewUnappliedTransactions(&date))

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 5, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	assert.Equal(suite.T(), "", results[0]["Entity"], "Entity - Unapplied payments Debit")
	assert.Equal(suite.T(), "", results[0]["Cost Centre"], "Cost Centre - Unapplied payments Debit")
	assert.Equal(suite.T(), "", results[0]["Account"], "Account - Unapplied payments Debit")
	assert.Equal(suite.T(), "", results[0]["Objective"], "Objective - Unapplied payments Debit")
	assert.Equal(suite.T(), "", results[0]["Analysis"], "Analysis - Unapplied payments Debit")
	assert.Equal(suite.T(), "", results[0]["Intercompany"], "Intercompany - Unapplied payments Debit")
	assert.Equal(suite.T(), "", results[0]["Spare"], "Spare - Unapplied payments Debit")
	assert.Equal(suite.T(), "27.34", results[0]["Debit"], "Debit - Unapplied payments Debit")
	assert.Equal(suite.T(), "", results[0]["Credit"], "Credit - Unapplied payments Debit")
	assert.Equal(suite.T(), fmt.Sprintf("Unapplied payments [%s]", yesterday.Date().Format("02/01/2006")), results[0]["Line description"], "Line description - Unapplied payments Debit")

	assert.Equal(suite.T(), "", results[1]["Entity"], "Entity - Unapplied payments Credit")
	assert.Equal(suite.T(), "", results[1]["Cost Centre"], "Cost Centre - Unapplied payments Credit")
	assert.Equal(suite.T(), "", results[1]["Account"], "Account - Unapplied payments Credit")
	assert.Equal(suite.T(), "", results[1]["Objective"], "Objective - Unapplied payments Credit")
	assert.Equal(suite.T(), "", results[1]["Analysis"], "Analysis - Unapplied payments Credit")
	assert.Equal(suite.T(), "", results[1]["Intercompany"], "Intercompany - Unapplied payments Credit")
	assert.Equal(suite.T(), "", results[1]["Spare"], "Spare - Unapplied payments Credit")
	assert.Equal(suite.T(), "", results[1]["Debit"], "Debit - Unapplied payments Credit")
	assert.Equal(suite.T(), "27.34", results[1]["Credit"], "Credit - Unapplied payments Credit")
	assert.Equal(suite.T(), fmt.Sprintf("Unapplied payments [%s]", yesterday.Date().Format("02/01/2006")), results[1]["Line description"], "Line description - Unapplied payments Credit")

	assert.Equal(suite.T(), "", results[2]["Entity"], "Entity - Reapplied payments Debit")
	assert.Equal(suite.T(), "", results[2]["Cost Centre"], "Cost Centre - Reapplied payments Debit")
	assert.Equal(suite.T(), "", results[2]["Account"], "Account - Reapplied payments Debit")
	assert.Equal(suite.T(), "", results[2]["Objective"], "Objective - Reapplied payments Debit")
	assert.Equal(suite.T(), "", results[2]["Analysis"], "Analysis - Reapplied payments Debit")
	assert.Equal(suite.T(), "", results[2]["Intercompany"], "Intercompany - Reapplied payments Debit")
	assert.Equal(suite.T(), "", results[2]["Spare"], "Spare - Reapplied payments Debit")
	assert.Equal(suite.T(), "22.34", results[2]["Debit"], "Debit - Reapplied payments Debit")
	assert.Equal(suite.T(), "", results[2]["Credit"], "Credit - Reapplied payments Debit")
	assert.Equal(suite.T(), fmt.Sprintf("Reapplied payments [%s]", yesterday.Date().Format("02/01/2006")), results[2]["Line description"], "Line description - Reapplied payments Debit")

	assert.Equal(suite.T(), "", results[3]["Entity"], "Entity - Reapplied payments Credit")
	assert.Equal(suite.T(), "", results[3]["Cost Centre"], "Cost Centre - Reapplied payments Credit")
	assert.Equal(suite.T(), "", results[3]["Account"], "Account - Reapplied payments Credit")
	assert.Equal(suite.T(), "", results[3]["Objective"], "Objective - Reapplied payments Credit")
	assert.Equal(suite.T(), "", results[3]["Analysis"], "Analysis - Reapplied payments Credit")
	assert.Equal(suite.T(), "", results[3]["Intercompany"], "Intercompany - Reapplied payments Credit")
	assert.Equal(suite.T(), "", results[3]["Spare"], "Spare - Reapplied payments Credit")
	assert.Equal(suite.T(), "", results[3]["Debit"], "Debit - Reapplied payments Credit")
	assert.Equal(suite.T(), "22.34", results[3]["Credit"], "Credit - Reapplied payments Credit")
	assert.Equal(suite.T(), fmt.Sprintf("Reapplied payments [%s]", yesterday.Date().Format("02/01/2006")), results[3]["Line description"], "Line description - Reapplied payments Credit")
}
