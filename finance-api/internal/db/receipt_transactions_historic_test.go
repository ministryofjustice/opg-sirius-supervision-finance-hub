package db

import (
	"fmt"
	"strconv"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_receipt_transactions_historic() {
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

	// an additional MOTO payment that is misapplied and added onto the correct client, leading to overpayment
	client5ID := suite.seeder.CreateClient(ctx, "Ernie", "Error", "55555555", "2314", "ACTIVE")
	_, _ = suite.seeder.CreateInvoice(ctx, client5ID, shared.InvoiceTypeS2, &general, twoMonthsAgo.StringPtr(), nil, nil, nil, twoMonthsAgo.StringPtr())
	suite.seeder.CreatePayment(ctx, 1234, twoMonthsAgo.Date(), "55555555", shared.TransactionTypeMotoCardPayment, twoMonthsAgo.Date(), 0)

	client6ID := suite.seeder.CreateClient(ctx, "Colette", "Correct", "66666666", "2314", "ACTIVE")
	_, _ = suite.seeder.CreateInvoice(ctx, client6ID, shared.InvoiceTypeS3, &minimal, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())
	suite.seeder.ReversePayment(ctx, "55555555", "66666666", "12.34", twoMonthsAgo.Date(), twoMonthsAgo.Date(), shared.TransactionTypeMotoCardPayment, yesterday.Date())

	// one client with an S2 invoice, two cheques payments for the same PIS number and one cheque payment for another PIS number
	client7ID := suite.seeder.CreateClient(ctx, "Gilgamesh", "Test", "77777777", "9999", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client7ID)
	_, _ = suite.seeder.CreateInvoice(ctx, client7ID, shared.InvoiceTypeS2, &general, twoMonthsAgo.StringPtr(), nil, nil, nil, twoMonthsAgo.StringPtr())

	pisNumber1 := int32(100023)
	pisNumber2 := int32(100024)
	suite.seeder.CreatePayment(ctx, 4020, yesterday.Date(), "77777777", shared.TransactionTypeSupervisionChequePayment, yesterday.Date(), pisNumber1)
	suite.seeder.CreatePayment(ctx, 1700, yesterday.Date(), "77777777", shared.TransactionTypeSupervisionChequePayment, yesterday.Date(), pisNumber1)
	suite.seeder.CreatePayment(ctx, 1500, yesterday.Date(), "77777777", shared.TransactionTypeSupervisionChequePayment, yesterday.Date(), pisNumber2)

	c := Client{suite.seeder.Conn}

	rows, err := c.Run(ctx, NewReceiptTransactionsHistoric(ReceiptTransactionsHistoricInput{Date: &shared.Date{Time: yesterday.Date()}}))

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 19, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	assert.Equal(suite.T(), "=\"0470\"", results[0]["Entity"], "Entity - MOTO card Payments Debit")
	assert.Equal(suite.T(), "99999999", results[0]["Cost Centre"], "Cost Centre - MOTO card Payments Debit")
	assert.Equal(suite.T(), "1841102050", results[0]["Account"], "Account - MOTO card Payments Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[0]["Objective"], "Objective - MOTO card Payments Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[0]["Analysis"], "Analysis - MOTO card Payments Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[0]["Intercompany"], "Intercompany - MOTO card Payments Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[0]["Spare"], "Spare - MOTO card Payments Debit")
	assert.Equal(suite.T(), "12.34", results[0]["Debit"], "Debit - MOTO card Payments Debit")
	assert.Equal(suite.T(), "12.34", results[0]["Credit"], "Credit - MOTO card Payments Debit")
	assert.Equal(suite.T(), fmt.Sprintf("MOTO card [%s]", twoMonthsAgo.Date().Format("02/01/2006")), results[0]["Line description"], "Line description - MOTO card Payments Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[1]["Entity"], "Entity - MOTO card Payments Credit")
	assert.Equal(suite.T(), "99999999", results[1]["Cost Centre"], "Cost Centre - MOTO card Payments Credit")
	assert.Equal(suite.T(), "1816102003", results[1]["Account"], "Account - MOTO card Payments Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[1]["Objective"], "Objective - MOTO card Payments Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[1]["Analysis"], "Analysis - MOTO card Payments Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[1]["Intercompany"], "Intercompany - MOTO card Payments Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[1]["Spare"], "Spare - MOTO card Payments Credit")
	assert.Equal(suite.T(), "12.34", results[1]["Debit"], "Debit - MOTO card Payments Credit")
	assert.Equal(suite.T(), "10.00", results[1]["Credit"], "Credit - MOTO card Payments Credit")
	assert.Equal(suite.T(), fmt.Sprintf("MOTO card [%s]", twoMonthsAgo.Date().Format("02/01/2006")), results[1]["Line description"], "Line description - MOTO card Payments Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[2]["Entity"], "Entity - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "", results[2]["Cost Centre"], "Cost Centre - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "", results[2]["Account"], "Account - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "", results[2]["Objective"], "Objective - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "", results[2]["Analysis"], "Analysis - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "", results[2]["Intercompany"], "Intercompany - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "", results[2]["Spare"], "Spare - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "", results[2]["Debit"], "Debit - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "2.34", results[2]["Credit"], "Credit - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), fmt.Sprintf("MOTO card [%s]", twoMonthsAgo.Date().Format("02/01/2006")), results[2]["Line description"], "Line description - MOTO card Payments Overpayment")

	assert.Equal(suite.T(), "=\"0470\"", results[3]["Entity"], "Entity - MOTO card Payments Debit")
	assert.Equal(suite.T(), "99999999", results[3]["Cost Centre"], "Cost Centre - MOTO card Payments Debit")
	assert.Equal(suite.T(), "1841102050", results[3]["Account"], "Account - MOTO card Payments Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[3]["Objective"], "Objective - MOTO card Payments Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[3]["Analysis"], "Analysis - MOTO card Payments Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[3]["Intercompany"], "Intercompany - MOTO card Payments Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[3]["Spare"], "Spare - MOTO card Payments Debit")
	assert.Equal(suite.T(), "130.00", results[3]["Debit"], "Debit - MOTO card Payments Debit")
	assert.Equal(suite.T(), "", results[3]["Credit"], "Credit - MOTO card Payments Debit")
	assert.Equal(suite.T(), fmt.Sprintf("MOTO card [%s]", yesterday.Date().Format("02/01/2006")), results[3]["Line description"], "Line description - MOTO card Payments Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[4]["Entity"], "Entity - MOTO card Payments Credit")
	assert.Equal(suite.T(), "99999999", results[4]["Cost Centre"], "Cost Centre - MOTO card Payments Credit")
	assert.Equal(suite.T(), "1816102003", results[4]["Account"], "Account - MOTO card Payments Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[4]["Objective"], "Objective - MOTO card Payments Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[4]["Analysis"], "Analysis - MOTO card Payments Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[4]["Intercompany"], "Intercompany - MOTO card Payments Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[4]["Spare"], "Spare - MOTO card Payments Credit")
	assert.Equal(suite.T(), "", results[4]["Debit"], "Debit - MOTO card Payments Credit")
	assert.Equal(suite.T(), "40.00", results[4]["Credit"], "Credit - MOTO card Payments Credit")
	assert.Equal(suite.T(), fmt.Sprintf("MOTO card [%s]", yesterday.Date().Format("02/01/2006")), results[4]["Line description"], "Line description - MOTO card Payments Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[5]["Entity"], "Entity - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "", results[5]["Cost Centre"], "Cost Centre - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "", results[5]["Account"], "Account - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "", results[5]["Objective"], "Objective - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "", results[5]["Analysis"], "Analysis - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "", results[5]["Intercompany"], "Intercompany - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "", results[5]["Spare"], "Spare - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "", results[5]["Debit"], "Debit - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), "90.00", results[5]["Credit"], "Credit - MOTO card Payments Overpayment")
	assert.Equal(suite.T(), fmt.Sprintf("MOTO card [%s]", yesterday.Date().Format("02/01/2006")), results[5]["Line description"], "Line description - MOTO card Payments Overpayment")

	assert.Equal(suite.T(), "=\"0470\"", results[6]["Entity"], "Entity - Online card Payments Debit")
	assert.Equal(suite.T(), "99999999", results[6]["Cost Centre"], "Cost Centre - Online card Payments Debit")
	assert.Equal(suite.T(), "1841102050", results[6]["Account"], "Account - Online card Payments Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[6]["Objective"], "Objective - Online card Payments Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[6]["Analysis"], "Analysis - Online card Payments Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[6]["Intercompany"], "Intercompany - Online card Payments Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[6]["Spare"], "Spare - Online card Payments Debit")
	assert.Equal(suite.T(), "17.00", results[6]["Debit"], "Debit - Online card Payments Debit")
	assert.Equal(suite.T(), "", results[6]["Credit"], "Credit - Online card Payments Debit")
	assert.Equal(suite.T(), fmt.Sprintf("Online card [%s]", yesterday.Date().Format("02/01/2006")), results[6]["Line description"], "Line description - Online card Payments Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[7]["Entity"], "Entity - Online card Payments Credit")
	assert.Equal(suite.T(), "99999999", results[7]["Cost Centre"], "Cost Centre - Online card Payments Credit")
	assert.Equal(suite.T(), "1816102003", results[7]["Account"], "Account - Online card Payments Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[7]["Objective"], "Objective - Online card Payments Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[7]["Analysis"], "Analysis - Online card Payments Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[7]["Intercompany"], "Intercompany - Online card Payments Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[7]["Spare"], "Spare - Online card Payments Credit")
	assert.Equal(suite.T(), "", results[7]["Debit"], "Debit - Online card Payments Credit")
	assert.Equal(suite.T(), "17.00", results[7]["Credit"], "Credit - Online card Payments Credit")
	assert.Equal(suite.T(), fmt.Sprintf("Online card [%s]", yesterday.Date().Format("02/01/2006")), results[7]["Line description"], "Line description - Online card Payments Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[8]["Entity"], "Entity - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "99999999", results[8]["Cost Centre"], "Cost Centre - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "1841102050", results[8]["Account"], "Account - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[8]["Objective"], "Objective - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[8]["Analysis"], "Analysis - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[8]["Intercompany"], "Intercompany - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[8]["Spare"], "Spare - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "1.20", results[8]["Debit"], "Debit - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "", results[8]["Credit"], "Credit - OPG BACS Payments Debit")
	assert.Equal(suite.T(), fmt.Sprintf("OPG BACS [%s]", yesterday.Date().Format("02/01/2006")), results[8]["Line description"], "Line description - OPG BACS Payments Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[9]["Entity"], "Entity - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "99999999", results[9]["Cost Centre"], "Cost Centre - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "1816102003", results[9]["Account"], "Account - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[9]["Objective"], "Objective - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[9]["Analysis"], "Analysis - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[9]["Intercompany"], "Intercompany - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[9]["Spare"], "Spare - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "", results[9]["Debit"], "Debit - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "1.20", results[9]["Credit"], "Credit - OPG BACS Payments Credit")
	assert.Equal(suite.T(), fmt.Sprintf("OPG BACS [%s]", yesterday.Date().Format("02/01/2006")), results[9]["Line description"], "Line description - OPG BACS Payments Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[10]["Entity"], "Entity - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "99999999", results[10]["Cost Centre"], "Cost Centre - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "1841102088", results[10]["Account"], "Account - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[10]["Objective"], "Objective - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[10]["Analysis"], "Analysis - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[10]["Intercompany"], "Intercompany - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[10]["Spare"], "Spare - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "25.50", results[10]["Debit"], "Debit - OPG BACS Payments Debit")
	assert.Equal(suite.T(), "", results[10]["Credit"], "Credit - OPG BACS Payments Debit")
	assert.Equal(suite.T(), fmt.Sprintf("Supervision BACS [%s]", yesterday.Date().Format("02/01/2006")), results[10]["Line description"], "Line description - OPG BACS Payments Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[11]["Entity"], "Entity - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "99999999", results[11]["Cost Centre"], "Cost Centre - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "1816102003", results[11]["Account"], "Account - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[11]["Objective"], "Objective - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[11]["Analysis"], "Analysis - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[11]["Intercompany"], "Intercompany - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[11]["Spare"], "Spare - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "", results[11]["Debit"], "Debit - OPG BACS Payments Credit")
	assert.Equal(suite.T(), "25.50", results[11]["Credit"], "Credit - OPG BACS Payments Credit")
	assert.Equal(suite.T(), fmt.Sprintf("Supervision BACS [%s]", yesterday.Date().Format("02/01/2006")), results[11]["Line description"], "Line description - OPG BACS Payments Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[12]["Entity"], "Entity - Direct Debit Debit")
	assert.Equal(suite.T(), "99999999", results[12]["Cost Centre"], "Cost Centre - Direct Debit Debit")
	assert.Equal(suite.T(), "1841102050", results[12]["Account"], "Account - Direct Debit Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[12]["Objective"], "Objective - Direct Debit Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[12]["Analysis"], "Analysis - Direct Debit Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[12]["Intercompany"], "Intercompany - Direct Debit Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[12]["Spare"], "Spare - Direct Debit Debit")
	assert.Equal(suite.T(), "40.20", results[12]["Debit"], "Debit - Direct Debit Debit")
	assert.Equal(suite.T(), "", results[12]["Credit"], "Credit - Direct Debit Debit")
	assert.Equal(suite.T(), fmt.Sprintf("Direct debit [%s]", yesterday.Date().Format("02/01/2006")), results[12]["Line description"], "Line description - Direct Debit Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[13]["Entity"], "Entity - Direct Debit Credit")
	assert.Equal(suite.T(), "99999999", results[13]["Cost Centre"], "Cost Centre - Direct Debit Credit")
	assert.Equal(suite.T(), "1816102003", results[13]["Account"], "Account - Direct Debit Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[13]["Objective"], "Objective - Direct Debit Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[13]["Analysis"], "Analysis - Direct Debit Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[13]["Intercompany"], "Intercompany - Direct Debit Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[13]["Spare"], "Spare - Direct Debit Credit")
	assert.Equal(suite.T(), "", results[13]["Debit"], "Debit - Direct Debit Credit")
	assert.Equal(suite.T(), "40.20", results[13]["Credit"], "Credit - Direct Debit Credit")
	assert.Equal(suite.T(), fmt.Sprintf("Direct debit [%s]", yesterday.Date().Format("02/01/2006")), results[13]["Line description"], "Line description - Direct Debit Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[14]["Entity"], "Entity - Cheques 1 & 2 Debit")
	assert.Equal(suite.T(), "99999999", results[14]["Cost Centre"], "Cost Centre - Cheques 1 & 2 Debit")
	assert.Equal(suite.T(), "1841102050", results[14]["Account"], "Account - Cheques 1 & 2 Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[14]["Objective"], "Objective - Cheques 1 & 2 Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[14]["Analysis"], "Analysis - Cheques 1 & 2 Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[14]["Intercompany"], "Intercompany - Cheques 1 & 2 Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[14]["Spare"], "Spare - Cheques 1 & 2 Debit")
	assert.Equal(suite.T(), "57.20", results[14]["Debit"], "Debit - Cheques 1 & 2 Debit")
	assert.Equal(suite.T(), "", results[14]["Credit"], "Credit - Cheques 1 & 2 Debit")
	assert.Equal(suite.T(), "Cheque payment [100023]", results[14]["Line description"], "Line description - Cheques 1 & 2 Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[15]["Entity"], "Entity - Cheques 1 & 2 Credit")
	assert.Equal(suite.T(), "99999999", results[15]["Cost Centre"], "Cost Centre - Cheques 1 & 2 Credit")
	assert.Equal(suite.T(), "1816102003", results[15]["Account"], "Account - Cheques 1 & 2 Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[15]["Objective"], "Objective - Cheques 1 & 2 Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[15]["Analysis"], "Analysis - Cheques 1 & 2 Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[15]["Intercompany"], "Intercompany - Cheques 1 & 2 Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[15]["Spare"], "Spare - Cheques 1 & 2 Credit")
	assert.Equal(suite.T(), "", results[15]["Debit"], "Debit - Cheques 1 & 2 Credit")
	assert.Equal(suite.T(), "57.20", results[15]["Credit"], "Credit - Cheques 1 & 2 Credit")
	assert.Equal(suite.T(), "Cheque payment [100023]", results[15]["Line description"], "Line description - Cheques 1 & 2 Credit")

	assert.Equal(suite.T(), "=\"0470\"", results[16]["Entity"], "Entity - Cheques 3 Debit")
	assert.Equal(suite.T(), "99999999", results[16]["Cost Centre"], "Cost Centre - Cheques 3 Debit")
	assert.Equal(suite.T(), "1841102050", results[16]["Account"], "Account - Cheques 3 Debit")
	assert.Equal(suite.T(), "=\"0000000\"", results[16]["Objective"], "Objective - Cheques 3 Debit")
	assert.Equal(suite.T(), "=\"00000000\"", results[16]["Analysis"], "Analysis - Cheques 3 Debit")
	assert.Equal(suite.T(), "=\"0000\"", results[16]["Intercompany"], "Intercompany - Cheques 3 Debit")
	assert.Equal(suite.T(), "=\"000000\"", results[16]["Spare"], "Spare - Cheques 3 Debit")
	assert.Equal(suite.T(), "15.00", results[16]["Debit"], "Debit - Cheques 3 Debit")
	assert.Equal(suite.T(), "", results[16]["Credit"], "Credit - Cheques 3 Debit")
	assert.Equal(suite.T(), "Cheque payment [100024]", results[16]["Line description"], "Line description - Cheques 3 Debit")

	assert.Equal(suite.T(), "=\"0470\"", results[17]["Entity"], "Entity - Cheques 3 Credit")
	assert.Equal(suite.T(), "99999999", results[17]["Cost Centre"], "Cost Centre - Cheques 3 Credit")
	assert.Equal(suite.T(), "1816102003", results[17]["Account"], "Account - Cheques 3 Credit")
	assert.Equal(suite.T(), "=\"0000000\"", results[17]["Objective"], "Objective - Cheques 3 Credit")
	assert.Equal(suite.T(), "=\"00000000\"", results[17]["Analysis"], "Analysis - Cheques 3 Credit")
	assert.Equal(suite.T(), "=\"0000\"", results[17]["Intercompany"], "Intercompany - Cheques 3 Credit")
	assert.Equal(suite.T(), "=\"00000\"", results[17]["Spare"], "Spare - Cheques 3 Credit")
	assert.Equal(suite.T(), "", results[17]["Debit"], "Debit - Cheques 3 Credit")
	assert.Equal(suite.T(), "15.00", results[17]["Credit"], "Credit - Cheques 3 Credit")
	assert.Equal(suite.T(), "Cheque payment [100024]", results[17]["Line description"], "Line description - Cheques 3 Debit")
}
