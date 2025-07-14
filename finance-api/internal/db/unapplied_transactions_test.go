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
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12345678", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client1ID, "ACTIVE")
	_, _ = suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, twoMonthsAgo.StringPtr())
	suite.seeder.CreatePayment(ctx, 1500, yesterday.Date(), "12345678", shared.TransactionTypeMotoCardPayment, yesterday.Date(), 0)
	_ = suite.seeder.CreateFeeReduction(ctx, client1ID, shared.FeeReductionTypeExemption, strconv.Itoa(yesterday.Date().Year()-1), 4, "", yesterday.Date())

	// an existing 100.00 unapply that has 10.00 reapplied
	client2ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "87654321", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client2ID, "ACTIVE")
	_, _ = suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS2, &general, threeYearsAgo.StringPtr(), nil, nil, nil, threeYearsAgo.StringPtr())
	suite.seeder.CreatePayment(ctx, 42000, threeYearsAgo.Date(), "87654321", shared.TransactionTypeOPGBACSPayment, threeYearsAgo.Date(), 0)
	_ = suite.seeder.CreateFeeReduction(ctx, client2ID, shared.FeeReductionTypeExemption, strconv.Itoa(threeYearsAgo.Date().Year()-1), 2, "", threeYearsAgo.Date())
	_, _ = suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS3, &minimal, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())

	// 12.34 unapply and reapply on the same day
	client3ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "33333333", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client3ID, "ACTIVE")
	invoiceID, _ := suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeS2, &general, twoMonthsAgo.StringPtr(), nil, nil, nil, twoMonthsAgo.StringPtr())
	suite.seeder.CreatePayment(ctx, 32000, yesterday.Date(), "33333333", shared.TransactionTypeMotoCardPayment, yesterday.Date(), 0)
	suite.seeder.CreateAdjustment(ctx, client3ID, invoiceID, shared.AdjustmentTypeCreditMemo, 1234, "Credit memo", yesterday.DatePtr())
	_, _ = suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeS2, &general, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())

	// 10.00 refund
	client4ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "44444444", "ACTIVE")
	suite.seeder.CreatePayment(ctx, 1000, yesterday.Date(), "44444444", shared.TransactionTypeMotoCardPayment, yesterday.Date(), 0)
	refundID := suite.seeder.CreateRefund(ctx, client4ID, "MR I TEST", "44444444", "44-44-44", yesterday.Date())
	suite.seeder.SetRefundDecision(ctx, client4ID, refundID, shared.RefundStatusApproved, yesterday.Date())

	suite.seeder.ProcessApprovedRefunds(ctx, []int32{refundID}, yesterday.Date())
	suite.seeder.FulfillRefund(ctx, refundID, 1000, yesterday.Date(), "44444444", "MR I TEST", "44444444", "444444", yesterday.Date())

	c := Client{suite.seeder.Conn}

	date := shared.NewDate(yesterday.String())

	rows, err := c.Run(ctx, NewUnappliedTransactions(UnappliedTransactionsInput{
		Date: &date,
	}))

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 7, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	assert.Equal(suite.T(), "", results[0]["Entity"], "Entity - Refunds Debit")
	assert.Equal(suite.T(), "", results[0]["Cost Centre"], "Cost Centre - Refunds Debit")
	assert.Equal(suite.T(), "", results[0]["Account"], "Account - Refunds Debit")
	assert.Equal(suite.T(), "", results[0]["Objective"], "Objective - Refunds Debit")
	assert.Equal(suite.T(), "", results[0]["Analysis"], "Analysis - Refunds Debit")
	assert.Equal(suite.T(), "", results[0]["Intercompany"], "Intercompany - Refunds Debit")
	assert.Equal(suite.T(), "", results[0]["Spare"], "Spare - Refunds Debit")
	assert.Equal(suite.T(), "10.00", results[0]["Debit"], "Debit - Refunds Debit")
	assert.Equal(suite.T(), "", results[0]["Credit"], "Credit - Refunds Debit")
	assert.Equal(suite.T(), fmt.Sprintf("Bankline refund [%s]", yesterday.Date().Format("02/01/2006")), results[0]["Line description"], "Line description - Refunds Debit")

	assert.Equal(suite.T(), "", results[1]["Entity"], "Entity - Refunds Credit")
	assert.Equal(suite.T(), "", results[1]["Cost Centre"], "Cost Centre - Refunds Credit")
	assert.Equal(suite.T(), "", results[1]["Account"], "Account - Refunds Credit")
	assert.Equal(suite.T(), "", results[1]["Objective"], "Objective - Refunds Credit")
	assert.Equal(suite.T(), "", results[1]["Analysis"], "Analysis - Refunds Credit")
	assert.Equal(suite.T(), "", results[1]["Intercompany"], "Intercompany - Refunds Credit")
	assert.Equal(suite.T(), "", results[1]["Spare"], "Spare - Refunds Credit")
	assert.Equal(suite.T(), "", results[1]["Debit"], "Debit - Refunds Credit")
	assert.Equal(suite.T(), "10.00", results[1]["Credit"], "Credit - Refunds Credit")
	assert.Equal(suite.T(), fmt.Sprintf("Bankline refund [%s]", yesterday.Date().Format("02/01/2006")), results[1]["Line description"], "Line description - Refunds Credit")

	assert.Equal(suite.T(), "", results[2]["Entity"], "Entity - Unapplied payments Debit")
	assert.Equal(suite.T(), "", results[2]["Cost Centre"], "Cost Centre - Unapplied payments Debit")
	assert.Equal(suite.T(), "", results[2]["Account"], "Account - Unapplied payments Debit")
	assert.Equal(suite.T(), "", results[2]["Objective"], "Objective - Unapplied payments Debit")
	assert.Equal(suite.T(), "", results[2]["Analysis"], "Analysis - Unapplied payments Debit")
	assert.Equal(suite.T(), "", results[2]["Intercompany"], "Intercompany - Unapplied payments Debit")
	assert.Equal(suite.T(), "", results[2]["Spare"], "Spare - Unapplied payments Debit")
	assert.Equal(suite.T(), "27.34", results[2]["Debit"], "Debit - Unapplied payments Debit")
	assert.Equal(suite.T(), "", results[2]["Credit"], "Credit - Unapplied payments Debit")
	assert.Equal(suite.T(), fmt.Sprintf("Unapplied payments [%s]", yesterday.Date().Format("02/01/2006")), results[2]["Line description"], "Line description - Unapplied payments Debit")

	assert.Equal(suite.T(), "", results[3]["Entity"], "Entity - Unapplied payments Credit")
	assert.Equal(suite.T(), "", results[3]["Cost Centre"], "Cost Centre - Unapplied payments Credit")
	assert.Equal(suite.T(), "", results[3]["Account"], "Account - Unapplied payments Credit")
	assert.Equal(suite.T(), "", results[3]["Objective"], "Objective - Unapplied payments Credit")
	assert.Equal(suite.T(), "", results[3]["Analysis"], "Analysis - Unapplied payments Credit")
	assert.Equal(suite.T(), "", results[3]["Intercompany"], "Intercompany - Unapplied payments Credit")
	assert.Equal(suite.T(), "", results[3]["Spare"], "Spare - Unapplied payments Credit")
	assert.Equal(suite.T(), "", results[3]["Debit"], "Debit - Unapplied payments Credit")
	assert.Equal(suite.T(), "27.34", results[3]["Credit"], "Credit - Unapplied payments Credit")
	assert.Equal(suite.T(), fmt.Sprintf("Unapplied payments [%s]", yesterday.Date().Format("02/01/2006")), results[3]["Line description"], "Line description - Unapplied payments Credit")

	assert.Equal(suite.T(), "", results[4]["Entity"], "Entity - Reapplied payments Debit")
	assert.Equal(suite.T(), "", results[4]["Cost Centre"], "Cost Centre - Reapplied payments Debit")
	assert.Equal(suite.T(), "", results[4]["Account"], "Account - Reapplied payments Debit")
	assert.Equal(suite.T(), "", results[4]["Objective"], "Objective - Reapplied payments Debit")
	assert.Equal(suite.T(), "", results[4]["Analysis"], "Analysis - Reapplied payments Debit")
	assert.Equal(suite.T(), "", results[4]["Intercompany"], "Intercompany - Reapplied payments Debit")
	assert.Equal(suite.T(), "", results[4]["Spare"], "Spare - Reapplied payments Debit")
	assert.Equal(suite.T(), "22.34", results[4]["Debit"], "Debit - Reapplied payments Debit")
	assert.Equal(suite.T(), "", results[4]["Credit"], "Credit - Reapplied payments Debit")
	assert.Equal(suite.T(), fmt.Sprintf("Reapplied payments [%s]", yesterday.Date().Format("02/01/2006")), results[4]["Line description"], "Line description - Reapplied payments Debit")

	assert.Equal(suite.T(), "", results[5]["Entity"], "Entity - Reapplied payments Credit")
	assert.Equal(suite.T(), "", results[5]["Cost Centre"], "Cost Centre - Reapplied payments Credit")
	assert.Equal(suite.T(), "", results[5]["Account"], "Account - Reapplied payments Credit")
	assert.Equal(suite.T(), "", results[5]["Objective"], "Objective - Reapplied payments Credit")
	assert.Equal(suite.T(), "", results[5]["Analysis"], "Analysis - Reapplied payments Credit")
	assert.Equal(suite.T(), "", results[5]["Intercompany"], "Intercompany - Reapplied payments Credit")
	assert.Equal(suite.T(), "", results[5]["Spare"], "Spare - Reapplied payments Credit")
	assert.Equal(suite.T(), "", results[5]["Debit"], "Debit - Reapplied payments Credit")
	assert.Equal(suite.T(), "22.34", results[5]["Credit"], "Credit - Reapplied payments Credit")
	assert.Equal(suite.T(), fmt.Sprintf("Reapplied payments [%s]", yesterday.Date().Format("02/01/2006")), results[5]["Line description"], "Line description - Reapplied payments Credit")
}
