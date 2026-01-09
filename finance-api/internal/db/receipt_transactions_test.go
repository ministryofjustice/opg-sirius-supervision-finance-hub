package db

import (
	"fmt"
	"strconv"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_receipt_transactions() {
	ctx := suite.ctx

	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)
	twoMonthsAgo := today.Sub(0, 2, 0)
	general := "320.00"
	minimal := "10.00"

	// one client with an invoice with a MOTO card payment, supervision BACS payment and an AD invoice with an exemption. This also creates an unapply.
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "11111111", "1234", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client1ID)
	_, _ = suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, twoMonthsAgo.StringPtr())

	suite.seeder.CreatePayment(ctx, 1500, yesterday.Date(), "11111111", shared.TransactionTypeMotoCardPayment, yesterday.Date(), 0)
	suite.seeder.CreatePayment(ctx, 2550, yesterday.Date(), "11111111", shared.TransactionTypeSupervisionBACSPayment, yesterday.Date(), 0)

	_ = suite.seeder.CreateFeeReduction(ctx, client1ID, shared.FeeReductionTypeExemption, strconv.Itoa(yesterday.Date().Year()-1), 4, "", yesterday.Date())

	// one client with an invoice with a MOTO card payment, an OPG BACS payment and an S2 invoice with an approved credit memo
	client2ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "22222222", "4321", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client2ID)
	invoice2ID, _ := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS2, &general, twoMonthsAgo.StringPtr(), nil, nil, nil, twoMonthsAgo.StringPtr())

	suite.seeder.CreatePayment(ctx, 120, yesterday.Date(), "22222222", shared.TransactionTypeOPGBACSPayment, yesterday.Date(), 0)
	suite.seeder.CreatePayment(ctx, 1500, yesterday.Date(), "22222222", shared.TransactionTypeMotoCardPayment, yesterday.Date(), 0)

	suite.seeder.CreateAdjustment(ctx, client2ID, invoice2ID, shared.AdjustmentTypeCreditMemo, -2500, "", yesterday.DatePtr())

	// one client with GA invoice, direct debit payment, online card payment
	client3ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "33333333", "2314", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client2ID)
	_, _ = suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeGA, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, twoMonthsAgo.StringPtr())

	suite.seeder.CreatePayment(ctx, 4020, yesterday.Date(), "33333333", shared.TransactionTypeDirectDebitPayment, yesterday.Date(), 0)
	suite.seeder.CreatePayment(ctx, 1700, yesterday.Date(), "33333333", shared.TransactionTypeOnlineCardPayment, yesterday.Date(), 0)

	// one client with two MOTO overpayments, one on a different date
	client4ID := suite.seeder.CreateClient(ctx, "Olive", "Overpayment", "44444444", "2314", "ACTIVE")
	_, _ = suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeS3, &minimal, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())
	suite.seeder.CreatePayment(ctx, 10000, yesterday.Date(), "44444444", shared.TransactionTypeMotoCardPayment, yesterday.Date(), 0)
	suite.seeder.CreatePayment(ctx, 10000, today.Date(), "44444444", shared.TransactionTypeMotoCardPayment, today.Date(), 0)

	// one client with an S2 invoice, two cheques payments for the same PIS number and one cheque payment for another PIS number
	client7ID := suite.seeder.CreateClient(ctx, "Gilgamesh", "Test", "77777777", "7777", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client7ID)
	_, _ = suite.seeder.CreateInvoice(ctx, client7ID, shared.InvoiceTypeS2, &general, twoMonthsAgo.StringPtr(), nil, nil, nil, twoMonthsAgo.StringPtr())

	pisNumber1 := int32(100023)
	pisNumber2 := int32(100024)
	suite.seeder.CreatePayment(ctx, 4020, yesterday.Date(), "77777777", shared.TransactionTypeSupervisionChequePayment, yesterday.Date(), pisNumber1)
	suite.seeder.CreatePayment(ctx, 1700, yesterday.Date(), "77777777", shared.TransactionTypeSupervisionChequePayment, yesterday.Date(), pisNumber1)
	suite.seeder.CreatePayment(ctx, 1500, yesterday.Date(), "77777777", shared.TransactionTypeSupervisionChequePayment, yesterday.Date(), pisNumber2)

	// refund - payment initially to credit balance
	client9ID := suite.seeder.CreateClient(ctx, "Randy", "Refund", "88888888", "1234", "ACTIVE")
	suite.seeder.CreatePayment(ctx, 14000, today.Date(), "88888888", shared.TransactionTypeMotoCardPayment, twoMonthsAgo.Date(), 0)
	refund4ID := suite.seeder.CreateRefund(ctx, client9ID, "MR R REFUND", "44444440", "44-44-44", twoMonthsAgo.Date())
	suite.seeder.SetRefundDecision(ctx, client9ID, refund4ID, shared.RefundStatusApproved, twoMonthsAgo.Date())

	suite.seeder.ProcessApprovedRefunds(ctx, []int32{refund4ID}, twoMonthsAgo.Date())
	suite.seeder.FulfillRefund(ctx, refund4ID, 14000, yesterday.Date(), "88888888", "MR R REFUND", "44444440", "444444", yesterday.Date())

	c := Client{suite.seeder.Conn}

	rows, err := c.Run(ctx, NewReceiptTransactions(ReceiptTransactionsInput{Date: &shared.Date{Time: yesterday.Date()}}))

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 18, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	assert.Equal(suite.T(), "=\"0470\"", results[0]["Entity"], "Entity - MOTO card Payments Debit")
	assert.Equal(suite.T(), "99999999", results[0]["Cost Centre"], "Cost Centre - MOTO card Payments Debit")
	assert.Equal(suite.T(), "1841102050", results[0]["Account"], "Account - MOTO card Payments Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[0]["Objective"], "Objective - MOTO card Payments Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[0]["Analysis"], "Analysis - MOTO card Payments Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[0]["Intercompany"], "Intercompany - MOTO card Payments Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[0]["Spare"], "Spare - MOTO card Payments Debit")
	assert.Equal(suite.T(), "130.00", results[0]["Debit"], "Debit - MOTO card Payments Debit")
	assert.Equal(suite.T(), "", results[0]["Credit"], "Credit - MOTO card Payments Debit")
	assert.Equal(suite.T(), fmt.Sprintf("MOTO card [%s]", yesterday.Date().Format("02/01/2006")), results[0]["Line description"], "Line description - MOTO card Payments Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[1]["Entity"], "Entity - MOTO card Payments Credit")
	assert.Equal(suite.T(), "99999999", results[1]["Cost Centre"], "Cost Centre - MOTO card Payments Credit")
	assert.Equal(suite.T(), "1816102003", results[1]["Account"], "Account - MOTO card Payments Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[1]["Objective"], "Objective - MOTO card Payments Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[1]["Analysis"], "Analysis - MOTO card Payments Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[1]["Intercompany"], "Intercompany - MOTO card Payments Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[1]["Spare"], "Spare - MOTO card Payments Credit")
	assert.Equal(suite.T(), "", results[1]["Debit"], "Debit - MOTO card Payments Credit")
	assert.Equal(suite.T(), "40.00", results[1]["Credit"], "Credit - MOTO card Payments Credit")
	assert.Equal(suite.T(), fmt.Sprintf("MOTO card [%s]", yesterday.Date().Format("02/01/2006")), results[1]["Line description"], "Line description - MOTO card Payments Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[2]["Entity"], "Entity - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "99999999", results[2]["Cost Centre"], "Cost Centre - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "1816102005", results[2]["Account"], "Account - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "=\"0000000\"", results[2]["Objective"], "Objective - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "=\"00000000\"", results[2]["Analysis"], "Analysis - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "=\"0000\"", results[2]["Intercompany"], "Intercompany - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "=\"00000\"", results[2]["Spare"], "Spare - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "", results[2]["Debit"], "Debit - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "90.00", results[2]["Credit"], "Credit - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), fmt.Sprintf("MOTO card [%s]", yesterday.Date().Format("02/01/2006")), results[2]["Line description"], "Line description - MOTO card Payments Overpayment")

	assert.Equal(suite.T(), "=\"0470\"", results[3]["Entity"], "Entity - Online card Payments Debit")
	assert.Equal(suite.T(), "99999999", results[3]["Cost Centre"], "Cost Centre - Online card Payments Debit")
	assert.Equal(suite.T(), "1841102050", results[3]["Account"], "Account - Online card Payments Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[3]["Objective"], "Objective - Online card Payments Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[3]["Analysis"], "Analysis - Online card Payments Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[3]["Intercompany"], "Intercompany - Online card Payments Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[3]["Spare"], "Spare - Online card Payments Debit")
	assert.Equal(suite.T(), "17.00", results[3]["Debit"], "Debit - Online card Payments Debit")
	assert.Equal(suite.T(), "", results[3]["Credit"], "Credit - Online card Payments Debit")
	assert.Equal(suite.T(), fmt.Sprintf("Online card [%s]", yesterday.Date().Format("02/01/2006")), results[3]["Line description"], "Line description - Online card Payments Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[4]["Entity"], "Entity - Online card Payments Credit")
	assert.Equal(suite.T(), "99999999", results[4]["Cost Centre"], "Cost Centre - Online card Payments Credit")
	assert.Equal(suite.T(), "1816102003", results[4]["Account"], "Account - Online card Payments Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[4]["Objective"], "Objective - Online card Payments Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[4]["Analysis"], "Analysis - Online card Payments Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[4]["Intercompany"], "Intercompany - Online card Payments Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[4]["Spare"], "Spare - Online card Payments Credit")
	assert.Equal(suite.T(), "", results[4]["Debit"], "Debit - Online card Payments Credit")
	assert.Equal(suite.T(), "17.00", results[4]["Credit"], "Credit - Online card Payments Credit")
	assert.Equal(suite.T(), fmt.Sprintf("Online card [%s]", yesterday.Date().Format("02/01/2006")), results[4]["Line description"], "Line description - Online card Payments Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[5]["Entity"], "Entity - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "99999999", results[5]["Cost Centre"], "Cost Centre - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "1841102050", results[5]["Account"], "Account - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[5]["Objective"], "Objective - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[5]["Analysis"], "Analysis - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[5]["Intercompany"], "Intercompany - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[5]["Spare"], "Spare - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "1.20", results[5]["Debit"], "Debit - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "", results[5]["Credit"], "Credit - OPG BACS Payments Debit")
	assert.Equal(suite.T(), fmt.Sprintf("OPG BACS [%s]", yesterday.Date().Format("02/01/2006")), results[5]["Line description"], "Line description - OPG BACS Payments Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[6]["Entity"], "Entity - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "99999999", results[6]["Cost Centre"], "Cost Centre - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "1816102003", results[6]["Account"], "Account - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[6]["Objective"], "Objective - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[6]["Analysis"], "Analysis - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[6]["Intercompany"], "Intercompany - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[6]["Spare"], "Spare - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "", results[6]["Debit"], "Debit - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "1.20", results[6]["Credit"], "Credit - OPG BACS Payments Credit")
	assert.Equal(suite.T(), fmt.Sprintf("OPG BACS [%s]", yesterday.Date().Format("02/01/2006")), results[6]["Line description"], "Line description - OPG BACS Payments Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[7]["Entity"], "Entity - Supervision BACS Payments Debit")
	assert.Equal(suite.T(), "99999999", results[7]["Cost Centre"], "Cost Centre - Supervision BACS Payments Debit")
	assert.Equal(suite.T(), "1841102088", results[7]["Account"], "Account - Supervision BACS Payments Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[7]["Objective"], "Objective - Supervision BACS Payments Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[7]["Analysis"], "Analysis - Supervision BACS Payments Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[7]["Intercompany"], "Intercompany - Supervision BACS Payments Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[7]["Spare"], "Spare - Supervision BACS Payments Debit")
	assert.Equal(suite.T(), "25.50", results[7]["Debit"], "Debit - Supervision BACS Payments Debit")
	assert.Equal(suite.T(), "", results[7]["Credit"], "Credit - Supervision BACS Payments Debit")
	assert.Equal(suite.T(), fmt.Sprintf("Supervision BACS [%s]", yesterday.Date().Format("02/01/2006")), results[7]["Line description"], "Line description - Supervision BACS Payments Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[8]["Entity"], "Entity - Supervision BACS Payments Credit")
	assert.Equal(suite.T(), "99999999", results[8]["Cost Centre"], "Cost Centre - Supervision BACS Payments Credit")
	assert.Equal(suite.T(), "1816102003", results[8]["Account"], "Account - Supervision BACS Payments Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[8]["Objective"], "Objective - Supervision BACS Payments Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[8]["Analysis"], "Analysis - Supervision BACS Payments Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[8]["Intercompany"], "Intercompany - Supervision BACS Payments Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[8]["Spare"], "Spare - Supervision BACS Payments Credit")
	assert.Equal(suite.T(), "", results[8]["Debit"], "Debit - Supervision BACS Payments Credit")
	assert.Equal(suite.T(), "25.50", results[8]["Credit"], "Credit - Supervision BACS Payments Credit")
	assert.Equal(suite.T(), fmt.Sprintf("Supervision BACS [%s]", yesterday.Date().Format("02/01/2006")), results[8]["Line description"], "Line description - Supervision BACS Payments Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[9]["Entity"], "Entity - Direct Debit Debit")
	assert.Equal(suite.T(), "99999999", results[9]["Cost Centre"], "Cost Centre - Direct Debit Debit")
	assert.Equal(suite.T(), "1841102050", results[9]["Account"], "Account - Direct Debit Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[9]["Objective"], "Objective - Direct Debit Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[9]["Analysis"], "Analysis - Direct Debit Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[9]["Intercompany"], "Intercompany - Direct Debit Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[9]["Spare"], "Spare - Direct Debit Debit")
	assert.Equal(suite.T(), "40.20", results[9]["Debit"], "Debit - Direct Debit Debit")
	assert.Equal(suite.T(), "", results[9]["Credit"], "Credit - Direct Debit Debit")
	assert.Equal(suite.T(), fmt.Sprintf("Direct debit [%s]", yesterday.Date().Format("02/01/2006")), results[9]["Line description"], "Line description - Direct Debit Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[10]["Entity"], "Entity - Direct Debit Credit")
	assert.Equal(suite.T(), "99999999", results[10]["Cost Centre"], "Cost Centre - Direct Debit Credit")
	assert.Equal(suite.T(), "1816102003", results[10]["Account"], "Account - Direct Debit Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[10]["Objective"], "Objective - Direct Debit Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[10]["Analysis"], "Analysis - Direct Debit Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[10]["Intercompany"], "Intercompany - Direct Debit Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[10]["Spare"], "Spare - Direct Debit Credit")
	assert.Equal(suite.T(), "", results[10]["Debit"], "Debit - Direct Debit Credit")
	assert.Equal(suite.T(), "40.20", results[10]["Credit"], "Credit - Direct Debit Credit")
	assert.Equal(suite.T(), fmt.Sprintf("Direct debit [%s]", yesterday.Date().Format("02/01/2006")), results[10]["Line description"], "Line description - Direct Debit Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[11]["Entity"], "Entity - Cheques 1 & 2 Debit")
	assert.Equal(suite.T(), "99999999", results[11]["Cost Centre"], "Cost Centre - Cheques 1 & 2 Debit")
	assert.Equal(suite.T(), "1841102050", results[11]["Account"], "Account - Cheques 1 & 2 Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[11]["Objective"], "Objective - Cheques 1 & 2 Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[11]["Analysis"], "Analysis - Cheques 1 & 2 Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[11]["Intercompany"], "Intercompany - Cheques 1 & 2 Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[11]["Spare"], "Spare - Cheques 1 & 2 Debit")
	assert.Equal(suite.T(), "57.20", results[11]["Debit"], "Debit - Cheques 1 & 2 Debit")
	assert.Equal(suite.T(), "", results[11]["Credit"], "Credit - Cheques 1 & 2 Debit")
	assert.Equal(suite.T(), "Cheque payment [100023]", results[11]["Line description"], "Line description - Cheques 1 & 2 Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[12]["Entity"], "Entity - Cheques 1 & 2 Credit")
	assert.Equal(suite.T(), "99999999", results[12]["Cost Centre"], "Cost Centre - Cheques 1 & 2 Credit")
	assert.Equal(suite.T(), "1816102003", results[12]["Account"], "Account - Cheques 1 & 2 Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[12]["Objective"], "Objective - Cheques 1 & 2 Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[12]["Analysis"], "Analysis - Cheques 1 & 2 Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[12]["Intercompany"], "Intercompany - Cheques 1 & 2 Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[12]["Spare"], "Spare - Cheques 1 & 2 Credit")
	assert.Equal(suite.T(), "", results[12]["Debit"], "Debit - Cheques 1 & 2 Credit")
	assert.Equal(suite.T(), "57.20", results[12]["Credit"], "Credit - Cheques 1 & 2 Credit")
	assert.Equal(suite.T(), "Cheque payment [100023]", results[12]["Line description"], "Line description - Cheques 1 & 2 Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[13]["Entity"], "Entity - Cheques 3 Debit")
	assert.Equal(suite.T(), "99999999", results[13]["Cost Centre"], "Cost Centre - Cheques 3 Debit")
	assert.Equal(suite.T(), "1841102050", results[13]["Account"], "Account - Cheques 3 Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[13]["Objective"], "Objective - Cheques 3 Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[13]["Analysis"], "Analysis - Cheques 3 Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[13]["Intercompany"], "Intercompany - Cheques 3 Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[13]["Spare"], "Spare - Cheques 3 Debit")
	assert.Equal(suite.T(), "15.00", results[13]["Debit"], "Debit - Cheques 3 Debit")
	assert.Equal(suite.T(), "", results[13]["Credit"], "Credit - Cheques 3 Debit")
	assert.Equal(suite.T(), "Cheque payment [100024]", results[13]["Line description"], "Line description - Cheques 3 Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[14]["Entity"], "Entity - Cheques 3 Credit")
	assert.Equal(suite.T(), "99999999", results[14]["Cost Centre"], "Cost Centre - Cheques 3 Credit")
	assert.Equal(suite.T(), "1816102003", results[14]["Account"], "Account - Cheques 3 Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[14]["Objective"], "Objective - Cheques 3 Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[14]["Analysis"], "Analysis - Cheques 3 Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[14]["Intercompany"], "Intercompany - Cheques 3 Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[14]["Spare"], "Spare - Cheques 3 Credit")
	assert.Equal(suite.T(), "", results[14]["Debit"], "Debit - Cheques 3 Credit")
	assert.Equal(suite.T(), "15.00", results[14]["Credit"], "Credit - Cheques 3 Credit")
	assert.Equal(suite.T(), "Cheque payment [100024]", results[14]["Line description"], "Line description - Cheques 3 Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[15]["Entity"], "Entity - Refund Debit")
	assert.Equal(suite.T(), "99999999", results[15]["Cost Centre"], "Cost Centre - Refund Debit")
	assert.Equal(suite.T(), "1841102050", results[15]["Account"], "Account - Refund Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[15]["Objective"], "Objective - Refund Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[15]["Analysis"], "Analysis - Refund Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[15]["Intercompany"], "Intercompany - Refund Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[15]["Spare"], "Spare - Refund Debit")
	assert.Equal(suite.T(), "", results[15]["Debit"], "Debit - Refund Debit")
	assert.Equal(suite.T(), "140.00", results[15]["Credit"], "Credit - Refund Debit")
	assert.Equal(suite.T(), fmt.Sprintf("Refund [%s]", yesterday.Date().Format("02/01/2006")), results[15]["Line description"], "Line description - Refund Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[16]["Entity"], "Entity - Refund Credit")
	assert.Equal(suite.T(), "99999999", results[16]["Cost Centre"], "Cost Centre - Refund Credit")
	assert.Equal(suite.T(), "1816102005", results[16]["Account"], "Account - Refund Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[16]["Objective"], "Objective - Refund Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[16]["Analysis"], "Analysis - Refund Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[16]["Intercompany"], "Intercompany - Refund Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[16]["Spare"], "Spare - Refund Credit")
	assert.Equal(suite.T(), "140.00", results[16]["Debit"], "Debit - Refund Credit")
	assert.Equal(suite.T(), "", results[16]["Credit"], "Credit - Refund Credit")
	assert.Equal(suite.T(), fmt.Sprintf("Refund [%s]", yesterday.Date().Format("02/01/2006")), results[16]["Line description"], "Line description - Refund Credit")
}

func (suite *IntegrationSuite) Test_receipt_transactions_reversals() {
	ctx := suite.ctx

	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)
	twoMonthsAgo := today.Sub(0, 2, 0)

	invoice1Amount := "120.00"
	invoice2Amount := "235.00"

	// Scenario 1: Client with invoice has misapplied cheque payment and overpayment, that is then reversed to the correct client, resulting in different overpayment
	client1ID := suite.seeder.CreateClient(ctx, "Misapplied", "Cheque", "11111111", "1111", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client1ID)
	_, _ = suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeS2, &invoice1Amount, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())
	pisNumber1 := int32(100001)
	suite.seeder.CreatePayment(ctx, 25000, yesterday.Date(), "11111111", shared.TransactionTypeSupervisionChequePayment, yesterday.Date(), pisNumber1)

	client2ID := suite.seeder.CreateClient(ctx, "Reversed", "Cheque", "22222222", "2222", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client2ID)
	_, _ = suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS2, &invoice2Amount, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())
	suite.seeder.ReversePayment(ctx, "11111111", "22222222", "250.00", yesterday.Date(), yesterday.Date(), shared.TransactionTypeSupervisionChequePayment, yesterday.Date(), "100001")

	// Scenario 2: Same as above but the original payment occurs in the past
	client3ID := suite.seeder.CreateClient(ctx, "Misapplied", "Cheque", "33333333", "3333", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client3ID)
	_, _ = suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeS2, &invoice1Amount, twoMonthsAgo.StringPtr(), nil, nil, nil, twoMonthsAgo.StringPtr())
	pisNumber2 := int32(100002)
	suite.seeder.CreatePayment(ctx, 25000, twoMonthsAgo.Date(), "33333333", shared.TransactionTypeSupervisionChequePayment, twoMonthsAgo.Date(), pisNumber2)

	client4ID := suite.seeder.CreateClient(ctx, "Reversed", "Cheque", "44444444", "4444", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client4ID)
	_, _ = suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeS2, &invoice2Amount, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())
	suite.seeder.ReversePayment(ctx, "33333333", "44444444", "250.00", twoMonthsAgo.Date(), twoMonthsAgo.Date(), shared.TransactionTypeSupervisionChequePayment, yesterday.Date(), "100002")

	// Scenario 3: reversed refund, invoice created between refund and reversal
	client10ID := suite.seeder.CreateClient(ctx, "Randy", "Reversal", "99999999", "1234", "ACTIVE")
	suite.seeder.CreatePayment(ctx, 14000, today.Date(), "99999999", shared.TransactionTypeMotoCardPayment, twoMonthsAgo.Date(), 0)
	refund5ID := suite.seeder.CreateRefund(ctx, client10ID, "MR R REVERSAL", "44444440", "44-44-44", twoMonthsAgo.Date())
	suite.seeder.SetRefundDecision(ctx, client10ID, refund5ID, shared.RefundStatusApproved, twoMonthsAgo.Date())

	suite.seeder.ProcessApprovedRefunds(ctx, []int32{refund5ID}, twoMonthsAgo.Date())
	suite.seeder.FulfillRefund(ctx, refund5ID, 14000, twoMonthsAgo.Date(), "99999999", "MR R REVERSAL", "44444440", "444444", twoMonthsAgo.Date())

	// refund occurred two months ago, reversal today
	_, _ = suite.seeder.CreateInvoice(ctx, client10ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())
	suite.seeder.ReverseRefund(ctx, "99999999", "140.00", twoMonthsAgo.Date(), yesterday.Date())

	c := Client{suite.seeder.Conn}

	rows, err := c.Run(ctx, NewReceiptTransactions(ReceiptTransactionsInput{Date: &shared.Date{Time: yesterday.Date()}}))

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 16, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// scenario 2: reversal of payment on a different day to the original payment (only reversal appears on journal)
	assert.Equal(suite.T(), "=\"0470\"", results[0]["Entity"], "Entity - Cheques 2 Debit")
	assert.Equal(suite.T(), "99999999", results[0]["Cost Centre"], "Cost Centre - Cheques 2 Debit")
	assert.Equal(suite.T(), "1841102050", results[0]["Account"], "Account - Cheques 2 Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[0]["Objective"], "Objective - Cheques 2 Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[0]["Analysis"], "Analysis - Cheques 2 Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[0]["Intercompany"], "Intercompany - Cheques 2 Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[0]["Spare"], "Spare - Cheques 2 Debit")
	assert.Equal(suite.T(), "250.00", results[0]["Debit"], "Debit - Cheques 2 Debit") // applied reversal of 250.00
	assert.Equal(suite.T(), "", results[0]["Credit"], "Credit - Cheques 2 Debit")
	assert.Equal(suite.T(), "Cheque payment [100002]", results[0]["Line description"], "Line description - Cheques 2 Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[1]["Entity"], "Entity - Cheques 2 Credit")
	assert.Equal(suite.T(), "99999999", results[1]["Cost Centre"], "Cost Centre - Cheques 2 Credit")
	assert.Equal(suite.T(), "1816102003", results[1]["Account"], "Account - Cheques 2 Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[1]["Objective"], "Objective - Cheques 2 Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[1]["Analysis"], "Analysis - Cheques 2 Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[1]["Intercompany"], "Intercompany - Cheques 2 Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[1]["Spare"], "Spare - Cheques 2 Credit")
	assert.Equal(suite.T(), "", results[1]["Debit"], "Debit - Cheques 2 Credit")
	assert.Equal(suite.T(), "235.00", results[1]["Credit"], "Credit - Cheques 2 Credit") // applied to debt
	assert.Equal(suite.T(), "Cheque payment [100002]", results[1]["Line description"], "Line description - Cheques 2 Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[2]["Entity"], "Entity - Cheques 2 Overpayment")
	assert.Equal(suite.T(), "99999999", results[2]["Cost Centre"], "Cost Centre - Cheques 2 Overpayment")
	assert.Equal(suite.T(), "1816102005", results[2]["Account"], "Account - Cheques 2 Overpayment")
	assert.Equal(suite.T(), "=\"0000000\"", results[2]["Objective"], "Objective - Cheques 2 Overpayment")
	assert.Equal(suite.T(), "=\"00000000\"", results[2]["Analysis"], "Analysis - Cheques 2 Overpayment")
	assert.Equal(suite.T(), "=\"0000\"", results[2]["Intercompany"], "Intercompany - Cheques 2 Overpayment")
	assert.Equal(suite.T(), "=\"00000\"", results[2]["Spare"], "Spare - Cheques 2 Overpayment")
	assert.Equal(suite.T(), "", results[2]["Debit"], "Debit - Cheques 2 Overpayment")
	assert.Equal(suite.T(), "15.00", results[2]["Credit"], "Credit - Cheques 2 Overpayment") // to CCB
	assert.Equal(suite.T(), "Cheque payment [100002]", results[2]["Line description"], "Line description - Cheques 2 Overpayment")

	assert.Equal(suite.T(), "=\"0470\"", results[3]["Entity"], "Entity - Cheques 2 Reversal Debit")
	assert.Equal(suite.T(), "99999999", results[3]["Cost Centre"], "Cost Centre - Cheques 2 Reversal Debit")
	assert.Equal(suite.T(), "1841102050", results[3]["Account"], "Account - Cheques 2 Reversal Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[3]["Objective"], "Objective - Cheques 2 Reversal Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[3]["Analysis"], "Analysis - Cheques 2 Reversal Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[3]["Intercompany"], "Intercompany - Cheques 2 Reversal Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[3]["Spare"], "Spare - Cheques 2 Reversal Debit")
	assert.Equal(suite.T(), "", results[3]["Debit"], "Debit - Cheques 2 Reversal Debit")
	assert.Equal(suite.T(), "250.00", results[3]["Credit"], "Credit - Cheques 2 Reversal Debit") // reversal of original payment
	assert.Equal(suite.T(), "Cheque payment [100002]", results[3]["Line description"], "Line description - Cheques 2 Reversal Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[4]["Entity"], "Entity - Cheques 2 Reversal Credit")
	assert.Equal(suite.T(), "99999999", results[4]["Cost Centre"], "Cost Centre - Cheques 2 Reversal Credit")
	assert.Equal(suite.T(), "1816102003", results[4]["Account"], "Account - Cheques 2 Reversal Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[4]["Objective"], "Objective - Cheques 2 Reversal Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[4]["Analysis"], "Analysis - Cheques 2 Reversal Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[4]["Intercompany"], "Intercompany - Cheques 2 Reversal Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[4]["Spare"], "Spare - Cheques 2 Reversal Credit")
	assert.Equal(suite.T(), "120.00", results[4]["Debit"], "Debit - Cheques 2 Reversal Credit") // reversal of original payment applied to invoice
	assert.Equal(suite.T(), "", results[4]["Credit"], "Credit - Cheques 2 Reversal Credit")
	assert.Equal(suite.T(), "Cheque payment [100002]", results[4]["Line description"], "Line description - Cheques 2 Reversal Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[5]["Entity"], "Entity - Cheques 2 Reversal Overpayment")
	assert.Equal(suite.T(), "99999999", results[5]["Cost Centre"], "Cost Centre - Cheques 2 Reversal Overpayment")
	assert.Equal(suite.T(), "1816102005", results[5]["Account"], "Account - Cheques 2 Reversal Overpayment")
	assert.Equal(suite.T(), "=\"0000000\"", results[5]["Objective"], "Objective - Cheques 2 Reversal Overpayment")
	assert.Equal(suite.T(), "=\"00000000\"", results[5]["Analysis"], "Analysis - Cheques 2 Reversal Overpayment")
	assert.Equal(suite.T(), "=\"0000\"", results[5]["Intercompany"], "Intercompany - Cheques 2 Reversal Overpayment")
	assert.Equal(suite.T(), "=\"00000\"", results[5]["Spare"], "Spare - Cheques 2 Reversal Overpayment")
	assert.Equal(suite.T(), "130.00", results[5]["Debit"], "Debit - Cheques 2 Reversal Overpayment") // reversal of original payment from credit
	assert.Equal(suite.T(), "", results[5]["Credit"], "Credit - Cheques 2 Reversal Overpayment")
	assert.Equal(suite.T(), "Cheque payment [100002]", results[5]["Line description"], "Line description - Cheques 2 Reversal Overpayment")

	// scenario 1: cheque payment misapplied and then reversed on the same day
	assert.Equal(suite.T(), "=\"0470\"", results[6]["Entity"], "Entity - Cheques 1 Debit")
	assert.Equal(suite.T(), "99999999", results[6]["Cost Centre"], "Cost Centre - Cheques 1 Debit")
	assert.Equal(suite.T(), "1841102050", results[6]["Account"], "Account - Cheques 1 Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[6]["Objective"], "Objective - Cheques 1 Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[6]["Analysis"], "Analysis - Cheques 1 Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[6]["Intercompany"], "Intercompany - Cheques 1 Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[6]["Spare"], "Spare - Cheques 1 Debit")
	assert.Equal(suite.T(), "500.00", results[6]["Debit"], "Debit - Cheques 1 Debit") // original payment of 250.00 + applied reversal of 250.00
	assert.Equal(suite.T(), "", results[6]["Credit"], "Credit - Cheques 1 Debit")
	assert.Equal(suite.T(), "Cheque payment [100001]", results[6]["Line description"], "Line description - Cheques 1 Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[7]["Entity"], "Entity - Cheques 1 Credit")
	assert.Equal(suite.T(), "99999999", results[7]["Cost Centre"], "Cost Centre - Cheques 1 Credit")
	assert.Equal(suite.T(), "1816102003", results[7]["Account"], "Account - Cheques 1 Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[7]["Objective"], "Objective - Cheques 1 Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[7]["Analysis"], "Analysis - Cheques 1 Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[7]["Intercompany"], "Intercompany - Cheques 1 Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[7]["Spare"], "Spare - Cheques 1 Credit")
	assert.Equal(suite.T(), "", results[7]["Debit"], "Debit - Cheques 1 Credit")
	assert.Equal(suite.T(), "355.00", results[7]["Credit"], "Credit - Cheques 1 Credit") // 120.00 + 235.00 on debt
	assert.Equal(suite.T(), "Cheque payment [100001]", results[7]["Line description"], "Line description - Cheques 1 Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[8]["Entity"], "Entity - Cheques 1 Overpayment")
	assert.Equal(suite.T(), "99999999", results[8]["Cost Centre"], "Cost Centre - Cheques 1 Overpayment")
	assert.Equal(suite.T(), "1816102005", results[8]["Account"], "Account - Cheques 1 Overpayment")
	assert.Equal(suite.T(), "=\"0000000\"", results[8]["Objective"], "Objective - Cheques 1 Overpayment")
	assert.Equal(suite.T(), "=\"00000000\"", results[8]["Analysis"], "Analysis - Cheques 1 Overpayment")
	assert.Equal(suite.T(), "=\"0000\"", results[8]["Intercompany"], "Intercompany - Cheques 1 Overpayment")
	assert.Equal(suite.T(), "=\"00000\"", results[8]["Spare"], "Spare - Cheques 1 Overpayment")
	assert.Equal(suite.T(), "", results[8]["Debit"], "Debit - Cheques 1 Overpayment")
	assert.Equal(suite.T(), "145.00", results[8]["Credit"], "Credit - Cheques 1 Overpayment") // 130 + 15 to CCB
	assert.Equal(suite.T(), "Cheque payment [100001]", results[8]["Line description"], "Line description - Cheques 1 Overpayment")

	assert.Equal(suite.T(), "=\"0470\"", results[9]["Entity"], "Entity - Cheques 1 Reversal Debit")
	assert.Equal(suite.T(), "99999999", results[9]["Cost Centre"], "Cost Centre - Cheques 1 Reversal Debit")
	assert.Equal(suite.T(), "1841102050", results[9]["Account"], "Account - Cheques 1 Reversal Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[9]["Objective"], "Objective - Cheques 1 Reversal Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[9]["Analysis"], "Analysis - Cheques 1 Reversal Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[9]["Intercompany"], "Intercompany - Cheques 1 Reversal Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[9]["Spare"], "Spare - Cheques 1 Reversal Debit")
	assert.Equal(suite.T(), "", results[9]["Debit"], "Debit - Cheques 1 Reversal Debit")
	assert.Equal(suite.T(), "250.00", results[9]["Credit"], "Credit - Cheques 1 Reversal Debit") // reversal of original payment
	assert.Equal(suite.T(), "Cheque payment [100001]", results[9]["Line description"], "Line description - Cheques 1 Reversal Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[10]["Entity"], "Entity - Cheques 1 Reversal Credit")
	assert.Equal(suite.T(), "99999999", results[10]["Cost Centre"], "Cost Centre - Cheques 1 Reversal Credit")
	assert.Equal(suite.T(), "1816102003", results[10]["Account"], "Account - Cheques 1 Reversal Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[10]["Objective"], "Objective - Cheques 1 Reversal Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[10]["Analysis"], "Analysis - Cheques 1 Reversal Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[10]["Intercompany"], "Intercompany - Cheques 1 Reversal Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[10]["Spare"], "Spare - Cheques 1 Reversal Credit")
	assert.Equal(suite.T(), "120.00", results[10]["Debit"], "Debit - Cheques 1 Reversal Credit") // reversal of original payment
	assert.Equal(suite.T(), "", results[10]["Credit"], "Credit - Cheques 1 Reversal Credit")
	assert.Equal(suite.T(), "Cheque payment [100001]", results[10]["Line description"], "Line description - Cheques 1 Reversal Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[11]["Entity"], "Entity - Cheques 1 Reversal Overpayment")
	assert.Equal(suite.T(), "99999999", results[11]["Cost Centre"], "Cost Centre - Cheques 1 Reversal Overpayment")
	assert.Equal(suite.T(), "1816102005", results[11]["Account"], "Account - Cheques 1 Reversal Overpayment")
	assert.Equal(suite.T(), "=\"0000000\"", results[11]["Objective"], "Objective - Cheques 1 Reversal Overpayment")
	assert.Equal(suite.T(), "=\"00000000\"", results[11]["Analysis"], "Analysis - Cheques 1 Reversal Overpayment")
	assert.Equal(suite.T(), "=\"0000\"", results[11]["Intercompany"], "Intercompany - Cheques 1 Reversal Overpayment")
	assert.Equal(suite.T(), "=\"00000\"", results[11]["Spare"], "Spare - Cheques 1 Reversal Overpayment")
	assert.Equal(suite.T(), "130.00", results[11]["Debit"], "Debit - Cheques 1 Reversal Overpayment")
	assert.Equal(suite.T(), "", results[11]["Credit"], "Credit - Cheques 1 Reversal Overpayment")
	assert.Equal(suite.T(), "Cheque payment [100001]", results[11]["Line description"], "Line description - Cheques 1 Reversal Overpayment")
	//1
	assert.Equal(suite.T(), "=\"0470\"", results[12]["Entity"], "Entity - Refund Reversal Debit")
	assert.Equal(suite.T(), "99999999", results[12]["Cost Centre"], "Cost Centre - Refund Reversal Debit")
	assert.Equal(suite.T(), "1841102050", results[12]["Account"], "Account - Refund Reversal Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[12]["Objective"], "Objective - Refund Reversal Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[12]["Analysis"], "Analysis - Refund Reversal Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[12]["Intercompany"], "Intercompany - Refund Reversal Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[12]["Spare"], "Spare - Refund Reversal Debit")
	assert.Equal(suite.T(), "140.00", results[12]["Debit"], "Debit - Refund Reversal Debit") // the other way round for refund reversals as these are effectively reversing a reversal
	assert.Equal(suite.T(), "", results[12]["Credit"], "Credit - Refund Reversal Debit")
	assert.Equal(suite.T(), fmt.Sprintf("Refund [%s]", yesterday.Date().Format("02/01/2006")), results[12]["Line description"], "Line description - Refund Reversal Debit")
	//2
	assert.Equal(suite.T(), "=\"0470\"", results[13]["Entity"], "Entity - Refund Reversal Credit")
	assert.Equal(suite.T(), "99999999", results[13]["Cost Centre"], "Cost Centre - Refund Reversal Credit")
	assert.Equal(suite.T(), "1816102003", results[13]["Account"], "Account - Refund Reversal Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[13]["Objective"], "Objective - Refund Reversal Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[13]["Analysis"], "Analysis - Refund Reversal Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[13]["Intercompany"], "Intercompany - Refund Reversal Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[13]["Spare"], "Spare - Refund Reversal Credit")
	assert.Equal(suite.T(), "", results[13]["Debit"], "Debit - Refund Reversal Credit")
	assert.Equal(suite.T(), "100.00", results[13]["Credit"], "Credit - Refund Reversal Credit") // the other way round for refund reversals as these are effectively reversing a reversal
	assert.Equal(suite.T(), fmt.Sprintf("Refund [%s]", yesterday.Date().Format("02/01/2006")), results[13]["Line description"], "Line description - Refund Reversal Credit")
	//3
	assert.Equal(suite.T(), "=\"0470\"", results[14]["Entity"], "Entity - Refund Reversal on invoice")
	assert.Equal(suite.T(), "99999999", results[14]["Cost Centre"], "Cost Centre - Refund Reversal on invoice")
	assert.Equal(suite.T(), "1816102005", results[14]["Account"], "Account - Refund Reversal on invoice")
	assert.Equal(suite.T(), "=\"0000000\"", results[14]["Objective"], "Objective - Refund Reversal on invoice")
	assert.Equal(suite.T(), "=\"00000000\"", results[14]["Analysis"], "Analysis - Refund Reversal on invoice")
	assert.Equal(suite.T(), "=\"0000\"", results[14]["Intercompany"], "Intercompany - Refund Reversal on invoice")
	assert.Equal(suite.T(), "=\"00000\"", results[14]["Spare"], "Spare - Refund Reversal on invoice")
	assert.Equal(suite.T(), "", results[14]["Debit"], "Debit - Refund Reversal on invoice")
	assert.Equal(suite.T(), "40.00", results[14]["Credit"], "Credit - Refund Reversal on invoice")
	assert.Equal(suite.T(), fmt.Sprintf("Refund [%s]", yesterday.Date().Format("02/01/2006")), results[14]["Line description"], "Line description - Refund Reversal on invoice")
}
