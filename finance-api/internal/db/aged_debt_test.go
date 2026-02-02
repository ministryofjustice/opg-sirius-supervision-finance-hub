package db

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/testhelpers"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_aged_debt() {
	ctx := suite.ctx
	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)
	twoMonthsAgo := today.Sub(0, 2, 0)
	elevenMonthsAgo := yesterday.Sub(0, 11, 0) // age will be ~0.917 years
	oneYearAgo := today.Sub(1, 0, 0)
	twoYearsAgo := today.Sub(2, 0, 0)
	fourYearsAgo := today.Sub(4, 0, 0)
	fiveYearsAgo := today.Sub(5, 0, 0)
	general := "320.00"

	// one client with:
	// - a lay deputy
	// - an active order
	// - one written off invoice
	// - one active invoice (2024)
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12345678", "1234", "ACTIVE")
	suite.seeder.CreateDeputy(ctx, client1ID, "Suzie", "Deputy", "LAY")
	suite.seeder.CreateOrder(ctx, client1ID)
	unpaidInvoiceID, c1i1Ref := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeGA, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, nil)
	paidInvoiceID, _ := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreateAdjustment(ctx, client1ID, paidInvoiceID, shared.AdjustmentTypeWriteOff, 0, "Written off", yesterday.DatePtr())
	// ignore these as legacy data with APPROVED ledger status
	suite.seeder.SeedData(
		fmt.Sprintf("INSERT INTO supervision_finance.ledger VALUES (99, 'ignore-me', '2022-04-11T08:36:40+00:00', '', 99999, '', 'CREDIT REMISSION', 'APPROVED', '%d', NULL, NULL, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 2);", client1ID),
		fmt.Sprintf("INSERT INTO supervision_finance.ledger_allocation VALUES (99, 99, '%d', '2022-04-11T08:36:40+00:00', 99999, 'ALLOCATED', NULL, 'Notes here', '2022-04-11', NULL);", unpaidInvoiceID),
		"ALTER SEQUENCE supervision_finance.ledger_id_seq RESTART WITH 100;",
		"ALTER SEQUENCE supervision_finance.ledger_allocation_id_seq RESTART WITH 100;",
	)

	// one client with:
	// - a pro deputy
	// - a closed order
	// - one active invoice (2020) with hardship reduction
	// - one active invoice (2022)
	client2ID := suite.seeder.CreateClient(ctx, "John", "Suite", "87654321", "4321", "ACTIVE")
	suite.seeder.CreateDeputy(ctx, client2ID, "Jane", "Deputy", "PRO")
	suite.seeder.CreateClosedOrder(ctx, client2ID, today.Date(), "")
	_ = suite.seeder.CreateFeeReduction(ctx, client2ID, shared.FeeReductionTypeRemission, strconv.Itoa(fiveYearsAgo.Date().Year()), 2, "A reduction", fiveYearsAgo.Date())
	_, c2i1Ref := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeAD, nil, fourYearsAgo.StringPtr(), nil, nil, nil, nil)
	_, c2i2Ref := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS2, &general, twoYearsAgo.StringPtr(), twoYearsAgo.StringPtr(), nil, nil, nil)

	// one client with:
	// split invoice
	i3amount := "170.00"
	client3ID := suite.seeder.CreateClient(ctx, "Freddy", "Splitz", "11111111", "1111", "ACTIVE")
	suite.seeder.CreateDeputy(ctx, client3ID, "Frank", "Deputy", "LAY")
	suite.seeder.CreateOrder(ctx, client3ID)
	c3i1ID, c3i1Ref := suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeS2, &i3amount, oneYearAgo.StringPtr(), oneYearAgo.StringPtr(), nil, nil, nil)
	suite.seeder.AddFeeRanges(ctx, c3i1ID, []testhelpers.FeeRange{
		{FromDate: oneYearAgo.Date(), ToDate: oneYearAgo.Add(0, 6, 0).Date(), SupervisionLevel: "GENERAL", Amount: 16000},
		{FromDate: oneYearAgo.Add(0, 6, 0).Date(), ToDate: oneYearAgo.Add(0, 11, 0).Date(), SupervisionLevel: "GENERAL", Amount: 1000},
	})

	// invoice paid off today but still included as debt, as received date is after to date
	client4ID := suite.seeder.CreateClient(ctx, "Penny", "Paid-Today", "44444444", "4444", "ACTIVE")
	suite.seeder.CreateDeputy(ctx, client4ID, "Franny", "Deputy", "LAY")
	suite.seeder.CreateOrder(ctx, client4ID)
	_, c4i1Ref := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeAD, nil, oneYearAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 10000, today.Date(), "44444444", shared.TransactionTypeMotoCardPayment, today.Date(), 0)

	// client with invoice ~11 months old (age between 0.9 and 1 year) to test edge case
	client5ID := suite.seeder.CreateClient(ctx, "Eddie", "Edge-Case", "55555555", "5555", "ACTIVE")
	suite.seeder.CreateDeputy(ctx, client5ID, "Emma", "Deputy", "LAY")
	suite.seeder.CreateOrder(ctx, client5ID)
	_, c5i1Ref := suite.seeder.CreateInvoice(ctx, client5ID, shared.InvoiceTypeAD, nil, elevenMonthsAgo.StringPtr(), nil, nil, nil, nil)

	c := Client{suite.seeder.Conn}

	to := shared.NewDate(yesterday.String())

	rows, err := c.Run(ctx, NewAgedDebt(AgedDebtInput{
		ToDate: &to,
		Today:  suite.seeder.Today().Add(1, 0, 0).Date(), // ran a year in the future to ensure data is independent of when it is generated
	}))

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 7, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// client 1
	assert.Equal(suite.T(), "Ian Test", results[0]["Customer name"], "Customer name - client 1")
	assert.Equal(suite.T(), "12345678", results[0]["Customer number"], "Customer number - client 1")
	assert.Equal(suite.T(), "1234", results[0]["SOP number"], "SOP number - client 1")
	assert.Equal(suite.T(), "LAY", results[0]["Deputy type"], "Deputy type - client 1")
	assert.Equal(suite.T(), "Yes", results[0]["Active case?"], "Active case? - client 1")
	assert.Equal(suite.T(), "=\"0470\"", results[0]["Entity"], "Entity - client 1")
	assert.Equal(suite.T(), "99999999", results[0]["Receivable cost centre"], "Receivable cost centre - client 1")
	assert.Equal(suite.T(), "BALANCE SHEET", results[0]["Receivable cost centre description"], "Receivable cost centre description - client 1")
	assert.Equal(suite.T(), "1816102003", results[0]["Receivable account code"], "Receivable account code - client 1")
	assert.Equal(suite.T(), "10486000", results[0]["Revenue cost centre"], "Revenue cost centre - client 1")
	assert.Equal(suite.T(), "Allocations, HW & SIS BISD", results[0]["Revenue cost centre description"], "Revenue cost centre description - client 1")
	assert.Equal(suite.T(), "4481102104", results[0]["Revenue account code"], "Revenue account code - client 1")
	assert.Equal(suite.T(), "INC - RECEIPT OF FEES AND CHARGES - GUARDIANSHIP ASSESS", results[0]["Revenue account code description"], "Revenue account code description - client 1")
	assert.Equal(suite.T(), "GA", results[0]["Invoice type"], "Invoice type - client 1")
	assert.Equal(suite.T(), c1i1Ref, results[0]["Trx number"], "Trx number - client 1")
	assert.Equal(suite.T(), "Guardianship assess invoice", results[0]["Transaction description"], "Transaction description - client 1")
	assert.Equal(suite.T(), twoMonthsAgo.String(), results[0]["Invoice date"], "Invoice date - client 1")
	assert.Equal(suite.T(), twoMonthsAgo.Add(0, 0, 30).String(), results[0]["Due date"], "Due date - client 1")
	assert.Equal(suite.T(), twoMonthsAgo.FinancialYear(), results[0]["Financial year"], "Financial year - client 1")
	assert.Equal(suite.T(), "30 NET", results[0]["Payment terms"], "Payment terms - client 1")
	assert.Equal(suite.T(), "200.00", results[0]["Original amount"], "Original amount - client 1")
	assert.Equal(suite.T(), "200.00", results[0]["Outstanding amount"], "Outstanding amount - client 1")
	assert.Equal(suite.T(), "0", results[0]["Current"], "Current - client 1")
	assert.Equal(suite.T(), "200.00", results[0]["0-1 years"], "0-1 years - client 1")
	assert.Equal(suite.T(), "0", results[0]["1-2 years"], "1-2 years - client 1")
	assert.Equal(suite.T(), "0", results[0]["2-3 years"], "2-3 years - client 1")
	assert.Equal(suite.T(), "0", results[0]["3-5 years"], "3-5 years - client 1")
	assert.Equal(suite.T(), "0", results[0]["5+ years"], "5+ years - client 1")
	assert.Equal(suite.T(), "=\"0-1\"", results[0]["Debt impairment years"], "Debt impairment years - client 1")

	// client 2 - invoice 2
	assert.Equal(suite.T(), "John Suite", results[1]["Customer name"], "Customer name - client 2, invoice 2")
	assert.Equal(suite.T(), "87654321", results[1]["Customer number"], "Customer number - client 2, invoice 2")
	assert.Equal(suite.T(), "4321", results[1]["SOP number"], "SOP number - client 2, invoice 2")
	assert.Equal(suite.T(), "PRO", results[1]["Deputy type"], "Deputy type - client 2, invoice 2")
	assert.Equal(suite.T(), "No", results[1]["Active case?"], "Active case? - client 2, invoice 2")
	assert.Equal(suite.T(), "=\"0470\"", results[1]["Entity"], "Entity - client 2, invoice 2")
	assert.Equal(suite.T(), "99999999", results[1]["Receivable cost centre"], "Receivable cost centre - client 2, invoice 2")
	assert.Equal(suite.T(), "BALANCE SHEET", results[1]["Receivable cost centre description"], "Receivable cost centre description - client 2, invoice 2")
	assert.Equal(suite.T(), "1816102003", results[1]["Receivable account code"], "Receivable account code - client 2, invoice 2")
	assert.Equal(suite.T(), "10482009", results[1]["Revenue cost centre"], "Revenue cost centre - client 2, invoice 2")
	assert.Equal(suite.T(), "Supervision Investigations", results[1]["Revenue cost centre description"], "Revenue cost centre description - client 2, invoice 2")
	assert.Equal(suite.T(), "4481102094", results[1]["Revenue account code"], "Revenue account code - client 2, invoice 2")
	assert.Equal(suite.T(), "INC - RECEIPT OF FEES AND CHARGES - Supervision Fee 1", results[1]["Revenue account code description"], "Revenue account code description - client 2, invoice 2")
	assert.Equal(suite.T(), "S2", results[1]["Invoice type"], "Invoice type - client 2, invoice 2")
	assert.Equal(suite.T(), c2i2Ref, results[1]["Trx number"], "Trx number - client 2, invoice 2")
	assert.Equal(suite.T(), "S2 - General invoice (Demanded)", results[1]["Transaction description"], "Transaction description - client 2, invoice 2")
	assert.Equal(suite.T(), twoYearsAgo.String(), results[1]["Invoice date"], "Invoice date - client 2, invoice 2")
	assert.Equal(suite.T(), twoYearsAgo.Add(0, 0, 30).String(), results[1]["Due date"], "Due date - client 2, invoice 2")
	assert.Equal(suite.T(), twoYearsAgo.FinancialYear(), results[1]["Financial year"], "Financial year - client 2, invoice 2")
	assert.Equal(suite.T(), "30 NET", results[1]["Payment terms"], "Payment terms - client 2, invoice 2")
	assert.Equal(suite.T(), "320.00", results[1]["Original amount"], "Original amount - client 2, invoice 2")
	assert.Equal(suite.T(), "320.00", results[1]["Outstanding amount"], "Outstanding amount - client 2, invoice 2")
	assert.Equal(suite.T(), "0", results[1]["Current"], "Current - client 2, invoice 2")
	assert.Equal(suite.T(), "0", results[1]["0-1 years"], "0-1 years - client 2, invoice 2")
	assert.Equal(suite.T(), "320.00", results[1]["1-2 years"], "1-2 years - client 2, invoice 2")
	assert.Equal(suite.T(), "0", results[1]["2-3 years"], "2-3 years - client 2, invoice 2")
	assert.Equal(suite.T(), "0", results[1]["3-5 years"], "3-5 years - client 2, invoice 2")
	assert.Equal(suite.T(), "0", results[1]["5+ years"], "5+ years - client 2, invoice 2")
	assert.Equal(suite.T(), "=\"3-5\"", results[1]["Debt impairment years"], "Debt impairment years - client 2, invoice 2")

	// client 2 - invoice 1
	assert.Equal(suite.T(), "John Suite", results[2]["Customer name"], "Customer name - client 2, invoice 1")
	assert.Equal(suite.T(), "87654321", results[2]["Customer number"], "Customer number - client 2, invoice 1")
	assert.Equal(suite.T(), "4321", results[2]["SOP number"], "SOP number - client 2, invoice 1")
	assert.Equal(suite.T(), "PRO", results[2]["Deputy type"], "Deputy type - client 2, invoice 1")
	assert.Equal(suite.T(), "No", results[2]["Active case?"], "Active case? - client 2, invoice 1")
	assert.Equal(suite.T(), "=\"0470\"", results[2]["Entity"], "Entity - client 2, invoice 1")
	assert.Equal(suite.T(), "99999999", results[2]["Receivable cost centre"], "Receivable cost centre - client 2, invoice 1")
	assert.Equal(suite.T(), "BALANCE SHEET", results[2]["Receivable cost centre description"], "Receivable cost centre description - client 2, invoice 1")
	assert.Equal(suite.T(), "1816102003", results[2]["Receivable account code"], "Receivable account code - client 2, invoice 1")
	assert.Equal(suite.T(), "10482009", results[2]["Revenue cost centre"], "Revenue cost centre - client 2, invoice 1")
	assert.Equal(suite.T(), "Supervision Investigations", results[2]["Revenue cost centre description"], "Revenue cost centre description - client 2, invoice 1")
	assert.Equal(suite.T(), "4481102093", results[2]["Revenue account code"], "Revenue account code - client 2, invoice 1")
	assert.Equal(suite.T(), "INC - RECEIPT OF FEES AND CHARGES - Appoint Deputy", results[2]["Revenue account code description"], "Revenue account code description - client 2, invoice 1")
	assert.Equal(suite.T(), "AD", results[2]["Invoice type"], "Invoice type - client 2, invoice 1")
	assert.Equal(suite.T(), c2i1Ref, results[2]["Trx number"], "Trx number - client 2, invoice 1")
	assert.Equal(suite.T(), "AD - Assessment deputy invoice", results[2]["Transaction description"], "Transaction description - client 2, invoice 1")
	assert.Equal(suite.T(), fourYearsAgo.String(), results[2]["Invoice date"], "Invoice date - client 2, invoice 1")
	assert.Equal(suite.T(), fourYearsAgo.Add(0, 0, 30).String(), results[2]["Due date"], "Due date - client 2, invoice 1")
	assert.Equal(suite.T(), fourYearsAgo.FinancialYear(), results[2]["Financial year"], "Financial year - client 2, invoice 1")
	assert.Equal(suite.T(), "30 NET", results[2]["Payment terms"], "Payment terms - client 2, invoice 1")
	assert.Equal(suite.T(), "100.00", results[2]["Original amount"], "Original amount - client 2, invoice 1")
	assert.Equal(suite.T(), "50.00", results[2]["Outstanding amount"], "Outstanding amount - client 2, invoice 1")
	assert.Equal(suite.T(), "0", results[2]["Current"], "Current - client 2, invoice 1")
	assert.Equal(suite.T(), "0", results[2]["0-1 years"], "0-1 years - client 2, invoice 1")
	assert.Equal(suite.T(), "0", results[2]["1-2 years"], "1-2 years - client 2, invoice 1")
	assert.Equal(suite.T(), "0", results[2]["2-3 years"], "2-3 years - client 2, invoice 1")
	assert.Equal(suite.T(), "50.00", results[2]["3-5 years"], "3-5 years - client 2, invoice 1")
	assert.Equal(suite.T(), "0", results[2]["5+ years"], "5+ years - client 2, invoice 1")
	assert.Equal(suite.T(), "=\"3-5\"", results[2]["Debt impairment years"], "Debt impairment years - client 2, invoice 1")

	assert.Equal(suite.T(), "Freddy Splitz", results[3]["Customer name"], "Customer Name - client 3")
	assert.Equal(suite.T(), "11111111", results[3]["Customer number"], "Customer number - client 3")
	assert.Equal(suite.T(), "1111", results[3]["SOP number"], "SOP number - client 3")
	assert.Equal(suite.T(), "LAY", results[3]["Deputy type"], "Deputy type - client 3")
	assert.Equal(suite.T(), "Yes", results[3]["Active case?"], "Active case? - client 3")
	assert.Equal(suite.T(), "=\"0470\"", results[3]["Entity"], "Entity - client 3")
	assert.Equal(suite.T(), "99999999", results[3]["Receivable cost centre"], "Receivable cost centre - client 3")
	assert.Equal(suite.T(), "BALANCE SHEET", results[3]["Receivable cost centre description"], "Receivable cost centre description - client 3")
	assert.Equal(suite.T(), "1816102003", results[3]["Receivable account code"], "Receivable account code - client 3")
	assert.Equal(suite.T(), "10482009", results[3]["Revenue cost centre"], "Revenue cost centre - client 3")
	assert.Equal(suite.T(), "Supervision Investigations", results[3]["Revenue cost centre description"], "Revenue cost centre description - client 3")
	assert.Equal(suite.T(), "4481102094", results[3]["Revenue account code"], "Revenue account code - client 3")
	assert.Equal(suite.T(), "INC - RECEIPT OF FEES AND CHARGES - Supervision Fee 1", results[3]["Revenue account code description"], "Revenue account code description - client 3")
	assert.Equal(suite.T(), "S2", results[3]["Invoice type"], "Invoice type - client 3")
	assert.Equal(suite.T(), c3i1Ref, results[3]["Trx number"], "Trx number - client 3")
	assert.Equal(suite.T(), "S2 - General invoice (Demanded)", results[3]["Transaction description"], "Transaction Description - client 3")
	assert.Equal(suite.T(), oneYearAgo.String(), results[3]["Invoice date"], "Invoice date - client 3")
	assert.Equal(suite.T(), oneYearAgo.Add(0, 0, 30).String(), results[3]["Due date"], "Due date - client 3")
	assert.Equal(suite.T(), oneYearAgo.FinancialYear(), results[3]["Financial year"], "Financial year - client 3")
	assert.Equal(suite.T(), "30 NET", results[3]["Payment terms"], "Payment terms - client 3")
	assert.Equal(suite.T(), "170.00", results[3]["Original amount"], "Original amount - client 3")
	assert.Equal(suite.T(), "170.00", results[3]["Outstanding amount"], "Outstanding amount - client 3")
	assert.Equal(suite.T(), "0", results[3]["Current"], "Current - client 3")
	assert.Equal(suite.T(), "170.00", results[3]["0-1 years"], "0-1 years - client 3")
	assert.Equal(suite.T(), "0", results[3]["1-2 years"], "1-2 years - client 3")
	assert.Equal(suite.T(), "0", results[3]["2-3 years"], "2-3 years - client 3")
	assert.Equal(suite.T(), "0", results[3]["3-5 years"], "3-5 years - client 3")
	assert.Equal(suite.T(), "0", results[3]["5+ years"], "5+ years - client 3")
	assert.Equal(suite.T(), "=\"0-1\"", results[3]["Debt impairment years"], "Debt impairment years - client 3")

	assert.Equal(suite.T(), "Penny Paid-Today", results[4]["Customer name"], "Customer Name - client 4")
	assert.Equal(suite.T(), "44444444", results[4]["Customer number"], "Customer number - client 4")
	assert.Equal(suite.T(), "4444", results[4]["SOP number"], "SOP number - client 4")
	assert.Equal(suite.T(), "LAY", results[4]["Deputy type"], "Deputy type - client 4")
	assert.Equal(suite.T(), "Yes", results[4]["Active case?"], "Active case? - client 4")
	assert.Equal(suite.T(), "=\"0470\"", results[4]["Entity"], "Entity - client 4")
	assert.Equal(suite.T(), "99999999", results[4]["Receivable cost centre"], "Receivable cost centre - client 4")
	assert.Equal(suite.T(), "BALANCE SHEET", results[4]["Receivable cost centre description"], "Receivable cost centre description - client 4")
	assert.Equal(suite.T(), "1816102003", results[4]["Receivable account code"], "Receivable account code - client 4")
	assert.Equal(suite.T(), "10482009", results[4]["Revenue cost centre"], "Revenue cost centre - client 4")
	assert.Equal(suite.T(), "Supervision Investigations", results[4]["Revenue cost centre description"], "Revenue cost centre description - client 4")
	assert.Equal(suite.T(), "4481102093", results[4]["Revenue account code"], "Revenue account code - client 4")
	assert.Equal(suite.T(), "INC - RECEIPT OF FEES AND CHARGES - Appoint Deputy", results[4]["Revenue account code description"], "Revenue account code description - client 4")
	assert.Equal(suite.T(), "AD", results[4]["Invoice type"], "Invoice type - client 4")
	assert.Equal(suite.T(), c4i1Ref, results[4]["Trx number"], "Trx number - client 4")
	assert.Equal(suite.T(), "AD - Assessment deputy invoice", results[4]["Transaction description"], "Transaction Description - client 4")
	assert.Equal(suite.T(), oneYearAgo.String(), results[4]["Invoice date"], "Invoice date - client 4")
	assert.Equal(suite.T(), oneYearAgo.Add(0, 0, 30).String(), results[4]["Due date"], "Due date - client 4")
	assert.Equal(suite.T(), oneYearAgo.FinancialYear(), results[4]["Financial year"], "Financial year - client 4")
	assert.Equal(suite.T(), "30 NET", results[4]["Payment terms"], "Payment terms - client 4")
	assert.Equal(suite.T(), "100.00", results[4]["Original amount"], "Original amount - client 4")
	assert.Equal(suite.T(), "100.00", results[4]["Outstanding amount"], "Outstanding amount - client 4")
	assert.Equal(suite.T(), "0", results[4]["Current"], "Current - client 4")
	assert.Equal(suite.T(), "100.00", results[4]["0-1 years"], "0-1 years - client 4")
	assert.Equal(suite.T(), "0", results[4]["1-2 years"], "1-2 years - client 4")
	assert.Equal(suite.T(), "0", results[4]["2-3 years"], "2-3 years - client 4")
	assert.Equal(suite.T(), "0", results[4]["3-5 years"], "3-5 years - client 4")
	assert.Equal(suite.T(), "0", results[4]["5+ years"], "5+ years - client 4")
	assert.Equal(suite.T(), "=\"0-1\"", results[4]["Debt impairment years"], "Debt impairment years - client 4")

	// client 5 - invoice with age ~0.917 years (11 months old)
	assert.Equal(suite.T(), "Eddie Edge-Case", results[5]["Customer name"], "Customer Name - client 5")
	assert.Equal(suite.T(), "55555555", results[5]["Customer number"], "Customer number - client 5")
	assert.Equal(suite.T(), "5555", results[5]["SOP number"], "SOP number - client 5")
	assert.Equal(suite.T(), "LAY", results[5]["Deputy type"], "Deputy type - client 5")
	assert.Equal(suite.T(), "Yes", results[5]["Active case?"], "Active case? - client 5")
	assert.Equal(suite.T(), "=\"0470\"", results[5]["Entity"], "Entity - client 5")
	assert.Equal(suite.T(), "99999999", results[5]["Receivable cost centre"], "Receivable cost centre - client 5")
	assert.Equal(suite.T(), "BALANCE SHEET", results[5]["Receivable cost centre description"], "Receivable cost centre description - client 5")
	assert.Equal(suite.T(), "1816102003", results[5]["Receivable account code"], "Receivable account code - client 5")
	assert.Equal(suite.T(), "10482009", results[5]["Revenue cost centre"], "Revenue cost centre - client 5")
	assert.Equal(suite.T(), "Supervision Investigations", results[5]["Revenue cost centre description"], "Revenue cost centre description - client 5")
	assert.Equal(suite.T(), "4481102093", results[5]["Revenue account code"], "Revenue account code - client 5")
	assert.Equal(suite.T(), "INC - RECEIPT OF FEES AND CHARGES - Appoint Deputy", results[5]["Revenue account code description"], "Revenue account code description - client 5")
	assert.Equal(suite.T(), "AD", results[5]["Invoice type"], "Invoice type - client 5")
	assert.Equal(suite.T(), c5i1Ref, results[5]["Trx number"], "Trx number - client 5")
	assert.Equal(suite.T(), "AD - Assessment deputy invoice", results[5]["Transaction description"], "Transaction Description - client 5")
	assert.Equal(suite.T(), elevenMonthsAgo.String(), results[5]["Invoice date"], "Invoice date - client 5")
	assert.Equal(suite.T(), elevenMonthsAgo.Add(0, 0, 30).String(), results[5]["Due date"], "Due date - client 5")
	assert.Equal(suite.T(), elevenMonthsAgo.FinancialYear(), results[5]["Financial year"], "Financial year - client 5")
	assert.Equal(suite.T(), "30 NET", results[5]["Payment terms"], "Payment terms - client 5")
	assert.Equal(suite.T(), "100.00", results[5]["Original amount"], "Original amount - client 5")
	assert.Equal(suite.T(), "100.00", results[5]["Outstanding amount"], "Outstanding amount - client 5")
	assert.Equal(suite.T(), "0", results[5]["Current"], "Current - client 5")
	assert.Equal(suite.T(), "100.00", results[5]["0-1 years"], "0-1 years - client 5")
	assert.Equal(suite.T(), "0", results[5]["1-2 years"], "1-2 years - client 5")
	assert.Equal(suite.T(), "0", results[5]["2-3 years"], "2-3 years - client 5")
	assert.Equal(suite.T(), "0", results[5]["3-5 years"], "3-5 years - client 5")
	assert.Equal(suite.T(), "0", results[5]["5+ years"], "5+ years - client 5")
	assert.Equal(suite.T(), "=\"0-1\"", results[5]["Debt impairment years"], "Debt impairment years - client 5")
}

func (suite *IntegrationSuite) Test_aged_debt_received_amount_considers_ledger_type_when_comparing_dates() {
	ctx := suite.ctx
	today := suite.seeder.Today()
	oneYearAgo := today.Sub(1, 0, 0).String()
	oneMonthAgo := today.Sub(0, 1, 0).String()
	twoWeeksAgo := today.Sub(0, 0, 14)
	twoWeeksAgoDate := shared.NewDate(twoWeeksAgo.String())
	oneWeekAgo := today.Sub(0, 0, 7)
	twoMonthsAgo := today.Sub(0, 2, 0)
	general := "320.00"
	tenYearsAgo := today.Sub(10, 0, 0).String()

	// non receipt type ledger will compare using the ledger date time
	// this ledger amount will be subtracted from the invoice amount on the report as the ledger date time (two months ago) is BEFORE the report run date
	cli1ID := suite.seeder.CreateClient(ctx, "Dan", "Testzilla", "44556677", "8258", "ACTIVE")
	unpaidInvoiceID3, cli1Ref := suite.seeder.CreateInvoice(ctx, cli1ID, shared.InvoiceTypeGA, &general, &oneYearAgo, nil, nil, nil, nil)
	suite.seeder.SeedData(
		fmt.Sprintf("INSERT INTO supervision_finance.ledger VALUES (1, 'ledger-ref1', '%s', '', 125, '', 'CREDIT MEMO', 'CONFIRMED', '%d', NULL, NULL, '%s', '%s', 1222, '', '', 1, '%s', 2);", twoMonthsAgo, cli1ID, twoMonthsAgo, twoMonthsAgo, oneMonthAgo),
		fmt.Sprintf("INSERT INTO supervision_finance.ledger_allocation VALUES (1, 1, '%d', '%s', 125, 'ALLOCATED', NULL, 'Notes here', '%s', NULL);", unpaidInvoiceID3, twoMonthsAgo, twoMonthsAgo),
	)

	// non receipt type ledger will compare using the ledger date time
	// this ledger amount will NOT be subtracted from the invoice amount on the report as the ledger date time (one week ago) is AFTER the report run date
	cli2ID := suite.seeder.CreateClient(ctx, "Ted", "Testington", "99551122", "4714", "ACTIVE")
	unpaidInvoiceID2, cli2Ref := suite.seeder.CreateInvoice(ctx, cli2ID, shared.InvoiceTypeGA, &general, &oneYearAgo, nil, nil, nil, nil)
	suite.seeder.SeedData(
		fmt.Sprintf("INSERT INTO supervision_finance.ledger VALUES (2, 'ledger-ref2', '%s', '', 125, '', 'CREDIT MEMO', 'CONFIRMED', '%d', NULL, NULL, '%s', '%s', 1222, '', '', 2, '%s', 2);", oneWeekAgo, cli2ID, oneWeekAgo, oneWeekAgo, oneMonthAgo),
		fmt.Sprintf("INSERT INTO supervision_finance.ledger_allocation VALUES (2, 2, '%d', '%s', 125, 'ALLOCATED', NULL, 'Notes here', '%s', NULL);", unpaidInvoiceID2, oneWeekAgo, oneWeekAgo),
	)

	// receipt type ledger generated before the finance hub go live date (01/04/25) will compare using the ledger created at time
	// this ledger amount will be subtracted from the invoice amount as the ledger datetime (ten years ago) is BEFORE the report run date
	cli3ID := suite.seeder.CreateClient(ctx, "Tod", "Testilla", "33445566", "5825", "ACTIVE")
	unpaidInvoiceID3, cli3Ref := suite.seeder.CreateInvoice(ctx, cli3ID, shared.InvoiceTypeGA, &general, &oneYearAgo, nil, nil, nil, nil)
	suite.seeder.SeedData(
		fmt.Sprintf("INSERT INTO supervision_finance.ledger VALUES (3, 'ledger-ref3', '%s', '', 300, '', 'OPG BACS PAYMENT', 'CONFIRMED', '%d', NULL, NULL, '%s', '%s', 1254, '', '', 3, null, 2);", tenYearsAgo, cli3ID, twoMonthsAgo, twoMonthsAgo),
		fmt.Sprintf("INSERT INTO supervision_finance.ledger_allocation VALUES (3, 3, '%d', '%s', 300, 'ALLOCATED', NULL, 'Notes here', '%s', NULL);", unpaidInvoiceID3, twoMonthsAgo, twoMonthsAgo),
	)

	// receipt type ledger generated after the finance hub go live date (01/04/25) will use the ledger created at date
	// this ledger amount will be subtracted from the invoice amount on the report as the ledger created at date (one month ago) is BEFORE the report run date (two weeks ago)
	cli4ID := suite.seeder.CreateClient(ctx, "Odin", "Testania", "55667788", "2582", "ACTIVE")
	unpaidInvoiceID4, cli4Ref := suite.seeder.CreateInvoice(ctx, cli4ID, shared.InvoiceTypeGA, &general, &oneYearAgo, nil, nil, nil, nil)
	suite.seeder.SeedData(
		fmt.Sprintf("INSERT INTO supervision_finance.ledger VALUES (4, 'ledger-ref-4', '%s', '', 125, '', 'OPG BACS PAYMENT', 'CONFIRMED', '%d', NULL, NULL, '%s', '%s', 1254, '', '', 4, '%s', 2);", twoMonthsAgo, cli4ID, twoMonthsAgo, twoMonthsAgo, oneMonthAgo),
		fmt.Sprintf("INSERT INTO supervision_finance.ledger_allocation VALUES (4, 4, '%d', '%s', 125, 'ALLOCATED', NULL, 'Notes here', '%s', NULL);", unpaidInvoiceID4, twoMonthsAgo, twoMonthsAgo),
	)

	// receipt type ledger generated after the finance hub go live date (01/04/25) will use the ledger created at date
	// this ledger amount will NOT be subtracted from the invoice amount on the report as the ledger created at date (one week ago) is AFTER the report run date (two weeks ago)
	cli5ID := suite.seeder.CreateClient(ctx, "Tanya", "Testilla", "33445566", "9963", "ACTIVE")
	unpaidInvoiceID5, cli5Ref := suite.seeder.CreateInvoice(ctx, cli5ID, shared.InvoiceTypeGA, &general, &oneYearAgo, nil, nil, nil, nil)
	suite.seeder.SeedData(
		fmt.Sprintf("INSERT INTO supervision_finance.ledger VALUES (5, 'ledger-ref', '%s', '', 123, '', 'OPG BACS PAYMENT', 'CONFIRMED', '%d', NULL, NULL, '%s', '%s', 1254, '', '', 5, '%s', 2);", twoMonthsAgo, cli5ID, twoMonthsAgo, twoMonthsAgo, oneWeekAgo),
		fmt.Sprintf("INSERT INTO supervision_finance.ledger_allocation VALUES (5, 5, '%d', '%s', 123, 'ALLOCATED', NULL, 'Notes here', '%s', NULL);", unpaidInvoiceID5, twoMonthsAgo, twoMonthsAgo),
	)

	c := Client{suite.seeder.Conn}

	rows, err := c.Run(ctx, NewAgedDebt(AgedDebtInput{
		ToDate: &twoWeeksAgoDate,
		Today:  suite.seeder.Today().Add(1, 0, 0).Date(), // ran a year in the future to ensure data is independent of when it is generated
	}))

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 6, len(rows))

	fmt.Print("Rows")
	fmt.Println(rows)

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// client 1 - non receipt - should apply
	assert.Equal(suite.T(), "Dan Testzilla", results[0]["Customer name"], "Customer name - client 1")
	assert.Equal(suite.T(), "44556677", results[0]["Customer number"], "Customer number - client 1")
	assert.Equal(suite.T(), "8258", results[0]["SOP number"], "SOP number - client 1")
	assert.Equal(suite.T(), cli1Ref, results[0]["Trx number"], "Trx number - client 1")
	assert.Equal(suite.T(), "200.00", results[0]["Original amount"], "Original amount - client 1")
	assert.Equal(suite.T(), "198.75", results[0]["Outstanding amount"], "Outstanding amount - client 1")
	assert.Equal(suite.T(), "0", results[0]["Current"], "Current - client 1")
	assert.Equal(suite.T(), "198.75", results[0]["0-1 years"], "0-1 years - client 1")
	assert.Equal(suite.T(), "0", results[0]["1-2 years"], "1-2 years - client 1")
	assert.Equal(suite.T(), "0", results[0]["2-3 years"], "2-3 years - client 1")
	assert.Equal(suite.T(), "0", results[0]["3-5 years"], "3-5 years - client 1")
	assert.Equal(suite.T(), "0", results[0]["5+ years"], "5+ years - client 1")

	// client 2 - non receipt - should not apply
	assert.Equal(suite.T(), "Ted Testington", results[1]["Customer name"], "Customer name - client 1")
	assert.Equal(suite.T(), "99551122", results[1]["Customer number"], "Customer number - client 1")
	assert.Equal(suite.T(), "4714", results[1]["SOP number"], "SOP number - client 1")
	assert.Equal(suite.T(), cli2Ref, results[1]["Trx number"], "Trx number - client 1")
	assert.Equal(suite.T(), "200.00", results[1]["Original amount"], "Original amount - client 1")
	assert.Equal(suite.T(), "200.00", results[1]["Outstanding amount"], "Outstanding amount - client 1")
	assert.Equal(suite.T(), "0", results[1]["Current"], "Current - client 1")
	assert.Equal(suite.T(), "200.00", results[1]["0-1 years"], "0-1 years - client 1")
	assert.Equal(suite.T(), "0", results[1]["1-2 years"], "1-2 years - client 1")
	assert.Equal(suite.T(), "0", results[1]["2-3 years"], "2-3 years - client 1")
	assert.Equal(suite.T(), "0", results[1]["3-5 years"], "3-5 years - client 1")
	assert.Equal(suite.T(), "0", results[1]["5+ years"], "5+ years - client 1")

	//client 3 - receipt before finance hub live - should apply
	assert.Equal(suite.T(), "Tod Testilla", results[2]["Customer name"], "Customer name - client 2")
	assert.Equal(suite.T(), "33445566", results[2]["Customer number"], "Customer number - client 2")
	assert.Equal(suite.T(), "5825", results[2]["SOP number"], "SOP number - client 2")
	assert.Equal(suite.T(), cli3Ref, results[2]["Trx number"], "Trx number - client 1")
	assert.Equal(suite.T(), "200.00", results[2]["Original amount"], "Original amount - client 1")
	assert.Equal(suite.T(), "197.00", results[2]["Outstanding amount"], "Outstanding amount - client 1")
	assert.Equal(suite.T(), "0", results[2]["Current"], "Current - client 1")
	assert.Equal(suite.T(), "197.00", results[2]["0-1 years"], "0-1 years - client 1")
	assert.Equal(suite.T(), "0", results[2]["1-2 years"], "1-2 years - client 1")
	assert.Equal(suite.T(), "0", results[2]["2-3 years"], "2-3 years - client 1")
	assert.Equal(suite.T(), "0", results[2]["3-5 years"], "3-5 years - client 1")
	assert.Equal(suite.T(), "0", results[2]["5+ years"], "5+ years - client 1")

	//client 4 - receipt after finance hub live - should apply
	assert.Equal(suite.T(), "Odin Testania", results[3]["Customer name"], "Customer name - client 1")
	assert.Equal(suite.T(), "55667788", results[3]["Customer number"], "Customer number - client 1")
	assert.Equal(suite.T(), "2582", results[3]["SOP number"], "SOP number - client 1")
	assert.Equal(suite.T(), cli4Ref, results[3]["Trx number"], "Trx number - client 1")
	assert.Equal(suite.T(), "200.00", results[3]["Original amount"], "Original amount - client 1")
	assert.Equal(suite.T(), "198.75", results[3]["Outstanding amount"], "Outstanding amount - client 1")
	assert.Equal(suite.T(), "0", results[3]["Current"], "Current - client 1")
	assert.Equal(suite.T(), "198.75", results[3]["0-1 years"], "0-1 years - client 1")
	assert.Equal(suite.T(), "0", results[3]["1-2 years"], "1-2 years - client 1")
	assert.Equal(suite.T(), "0", results[3]["2-3 years"], "2-3 years - client 1")
	assert.Equal(suite.T(), "0", results[3]["3-5 years"], "3-5 years - client 1")
	assert.Equal(suite.T(), "0", results[3]["5+ years"], "5+ years - client 1")

	//client 4 - receipt after finance hub live - should not apply
	assert.Equal(suite.T(), "Tanya Testilla", results[4]["Customer name"], "Customer name - client 1")
	assert.Equal(suite.T(), "33445566", results[4]["Customer number"], "Customer number - client 1")
	assert.Equal(suite.T(), "9963", results[4]["SOP number"], "SOP number - client 1")
	assert.Equal(suite.T(), cli5Ref, results[4]["Trx number"], "Trx number - client 1")
	assert.Equal(suite.T(), "200.00", results[4]["Original amount"], "Original amount - client 1")
	assert.Equal(suite.T(), "200.00", results[4]["Outstanding amount"], "Outstanding amount - client 1")
	assert.Equal(suite.T(), "0", results[4]["Current"], "Current - client 1")
	assert.Equal(suite.T(), "200.00", results[4]["0-1 years"], "0-1 years - client 1")
	assert.Equal(suite.T(), "0", results[4]["1-2 years"], "1-2 years - client 1")
	assert.Equal(suite.T(), "0", results[4]["2-3 years"], "2-3 years - client 1")
	assert.Equal(suite.T(), "0", results[4]["3-5 years"], "3-5 years - client 1")
	assert.Equal(suite.T(), "0", results[4]["5+ years"], "5+ years - client 1")
}

func TestAgedDebt_GetParams(t *testing.T) {
	type fields struct {
		ReportQuery   ReportQuery
		AgedDebtInput AgedDebtInput
	}
	tests := []struct {
		name   string
		fields fields
		want   []any
	}{
		{
			name:   "nil ToDate defaults to today",
			fields: fields{AgedDebtInput: AgedDebtInput{ToDate: nil, Today: time.Now()}},
			want:   []any{time.Now().Format("2006-01-02"), shared.NewDate("2020-01-01").Time},
		},
		{
			name:   "empty ToDate defaults to today",
			fields: fields{AgedDebtInput: AgedDebtInput{ToDate: &shared.Date{}, Today: time.Now()}},
			want:   []any{time.Now().Format("2006-01-02"), shared.NewDate("01/01/2020").Time},
		},
		{
			name:   "will pull through other date to overwrite today and default date if required",
			fields: fields{AgedDebtInput: AgedDebtInput{ToDate: &shared.Date{}, Today: time.Now().AddDate(-1, 0, 0)}},
			want:   []any{time.Now().AddDate(-1, 0, 0).Format("2006-01-02"), shared.NewDate("01/01/2020").Time},
		},
		{
			name: "valid ToDate returns formatted date",
			fields: fields{AgedDebtInput: AgedDebtInput{ToDate: func() *shared.Date {
				d := shared.NewDate("2023-05-01")
				return &d
			}(), Today: time.Now()}},
			want: []any{"2023-05-01", shared.NewDate("01/01/2020").Time},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AgedDebt{
				ReportQuery:   tt.fields.ReportQuery,
				AgedDebtInput: tt.fields.AgedDebtInput,
			}
			assert.Equalf(t, tt.want, a.GetParams(), "GetParams()")
		})
	}
}
