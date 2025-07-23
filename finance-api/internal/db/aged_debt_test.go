package db

import (
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/testhelpers"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"strconv"
)

func (suite *IntegrationSuite) Test_aged_debt() {
	ctx := suite.ctx
	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)
	twoMonthsAgo := today.Sub(0, 2, 0)
	oneYearAgo := today.Sub(1, 0, 0)
	twoYearsAgo := today.Sub(2, 0, 0)
	fourYearsAgo := today.Sub(4, 0, 0)
	fiveYearsAgo := today.Sub(5, 0, 0)
	sixYearsAgo := today.Sub(6, 0, 0)
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
	suite.seeder.CreateAdjustment(ctx, client1ID, paidInvoiceID, shared.AdjustmentTypeWriteOff, 0, "Written off", nil)
	// ignore these as legacy data with APPROVED ledger status
	suite.seeder.SeedData(
		fmt.Sprintf("INSERT INTO supervision_finance.ledger VALUES (99, 'ignore-me', '2022-04-11T08:36:40+00:00', '', 99999, '', 'CREDIT REMISSION', 'APPROVED', %d, NULL, NULL, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 2);", client1ID),
		fmt.Sprintf("INSERT INTO supervision_finance.ledger_allocation VALUES (99, 99, %d, '2022-04-11T08:36:40+00:00', 99999, 'ALLOCATED', NULL, 'Notes here', '2022-04-11', NULL);", unpaidInvoiceID),
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

	// excluded clients as out of range
	excluded1ID := suite.seeder.CreateClient(ctx, "Too", "Early", "99999999", "9999", "ACTIVE")
	suite.seeder.CreateInvoice(ctx, excluded1ID, shared.InvoiceTypeAD, nil, sixYearsAgo.StringPtr(), nil, nil, nil, nil)
	excluded2ID := suite.seeder.CreateClient(ctx, "Too", "Early", "99999999", "", "ACTIVE")
	suite.seeder.CreateInvoice(ctx, excluded2ID, shared.InvoiceTypeAD, nil, today.StringPtr(), nil, nil, nil, nil)

	c := Client{suite.seeder.Conn}

	from := shared.NewDate(fourYearsAgo.String())
	to := shared.NewDate(yesterday.String())

	rows, err := c.Run(ctx, NewAgedDebt(AgedDebtInput{
		FromDate: &from,
		ToDate:   &to,
	}))

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 5, len(rows))

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

	// client 2 - invoice 1
	assert.Equal(suite.T(), "John Suite", results[1]["Customer name"], "Customer name - client 2, invoice 1")
	assert.Equal(suite.T(), "87654321", results[1]["Customer number"], "Customer number - client 2, invoice 1")
	assert.Equal(suite.T(), "4321", results[1]["SOP number"], "SOP number - client 2, invoice 1")
	assert.Equal(suite.T(), "PRO", results[1]["Deputy type"], "Deputy type - client 2, invoice 1")
	assert.Equal(suite.T(), "No", results[1]["Active case?"], "Active case? - client 2, invoice 1")
	assert.Equal(suite.T(), "=\"0470\"", results[1]["Entity"], "Entity - client 2, invoice 1")
	assert.Equal(suite.T(), "99999999", results[1]["Receivable cost centre"], "Receivable cost centre - client 2, invoice 1")
	assert.Equal(suite.T(), "BALANCE SHEET", results[1]["Receivable cost centre description"], "Receivable cost centre description - client 2, invoice 1")
	assert.Equal(suite.T(), "1816102003", results[1]["Receivable account code"], "Receivable account code - client 2, invoice 1")
	assert.Equal(suite.T(), "10482009", results[1]["Revenue cost centre"], "Revenue cost centre - client 2, invoice 1")
	assert.Equal(suite.T(), "Supervision Investigations", results[1]["Revenue cost centre description"], "Revenue cost centre description - client 2, invoice 1")
	assert.Equal(suite.T(), "4481102093", results[1]["Revenue account code"], "Revenue account code - client 2, invoice 1")
	assert.Equal(suite.T(), "INC - RECEIPT OF FEES AND CHARGES - Appoint Deputy", results[1]["Revenue account code description"], "Revenue account code description - client 2, invoice 1")
	assert.Equal(suite.T(), "AD", results[1]["Invoice type"], "Invoice type - client 2, invoice 1")
	assert.Equal(suite.T(), c2i1Ref, results[1]["Trx number"], "Trx number - client 2, invoice 1")
	assert.Equal(suite.T(), "AD - Assessment deputy invoice", results[1]["Transaction description"], "Transaction description - client 2, invoice 1")
	assert.Equal(suite.T(), fourYearsAgo.String(), results[1]["Invoice date"], "Invoice date - client 2, invoice 1")
	assert.Equal(suite.T(), fourYearsAgo.Add(0, 0, 30).String(), results[1]["Due date"], "Due date - client 2, invoice 1")
	assert.Equal(suite.T(), fourYearsAgo.FinancialYear(), results[1]["Financial year"], "Financial year - client 2, invoice 1")
	assert.Equal(suite.T(), "30 NET", results[1]["Payment terms"], "Payment terms - client 2, invoice 1")
	assert.Equal(suite.T(), "100.00", results[1]["Original amount"], "Original amount - client 2, invoice 1")
	assert.Equal(suite.T(), "50.00", results[1]["Outstanding amount"], "Outstanding amount - client 2, invoice 1")
	assert.Equal(suite.T(), "0", results[1]["Current"], "Current - client 2, invoice 1")
	assert.Equal(suite.T(), "0", results[1]["0-1 years"], "0-1 years - client 2, invoice 1")
	assert.Equal(suite.T(), "0", results[1]["1-2 years"], "1-2 years - client 2, invoice 1")
	assert.Equal(suite.T(), "0", results[1]["2-3 years"], "2-3 years - client 2, invoice 1")
	assert.Equal(suite.T(), "50.00", results[1]["3-5 years"], "3-5 years - client 2, invoice 1")
	assert.Equal(suite.T(), "0", results[1]["5+ years"], "5+ years - client 2, invoice 1")
	assert.Equal(suite.T(), "=\"3-5\"", results[1]["Debt impairment years"], "Debt impairment years - client 2, invoice 1")

	// client 2 - invoice 2
	assert.Equal(suite.T(), "John Suite", results[2]["Customer name"], "Customer name - client 2, invoice 2")
	assert.Equal(suite.T(), "87654321", results[2]["Customer number"], "Customer number - client 2, invoice 2")
	assert.Equal(suite.T(), "4321", results[2]["SOP number"], "SOP number - client 2, invoice 2")
	assert.Equal(suite.T(), "PRO", results[2]["Deputy type"], "Deputy type - client 2, invoice 2")
	assert.Equal(suite.T(), "No", results[2]["Active case?"], "Active case? - client 2, invoice 2")
	assert.Equal(suite.T(), "=\"0470\"", results[2]["Entity"], "Entity - client 2, invoice 2")
	assert.Equal(suite.T(), "99999999", results[2]["Receivable cost centre"], "Receivable cost centre - client 2, invoice 2")
	assert.Equal(suite.T(), "BALANCE SHEET", results[2]["Receivable cost centre description"], "Receivable cost centre description - client 2, invoice 2")
	assert.Equal(suite.T(), "1816102003", results[2]["Receivable account code"], "Receivable account code - client 2, invoice 2")
	assert.Equal(suite.T(), "10482009", results[2]["Revenue cost centre"], "Revenue cost centre - client 2, invoice 2")
	assert.Equal(suite.T(), "Supervision Investigations", results[2]["Revenue cost centre description"], "Revenue cost centre description - client 2, invoice 2")
	assert.Equal(suite.T(), "4481102094", results[2]["Revenue account code"], "Revenue account code - client 2, invoice 2")
	assert.Equal(suite.T(), "INC - RECEIPT OF FEES AND CHARGES - Supervision Fee 1", results[2]["Revenue account code description"], "Revenue account code description - client 2, invoice 2")
	assert.Equal(suite.T(), "S2", results[2]["Invoice type"], "Invoice type - client 2, invoice 2")
	assert.Equal(suite.T(), c2i2Ref, results[2]["Trx number"], "Trx number - client 2, invoice 2")
	assert.Equal(suite.T(), "S2 - General invoice (Demanded)", results[2]["Transaction description"], "Transaction description - client 2, invoice 2")
	assert.Equal(suite.T(), twoYearsAgo.String(), results[2]["Invoice date"], "Invoice date - client 2, invoice 2")
	assert.Equal(suite.T(), twoYearsAgo.Add(0, 0, 30).String(), results[2]["Due date"], "Due date - client 2, invoice 2")
	assert.Equal(suite.T(), twoYearsAgo.FinancialYear(), results[2]["Financial year"], "Financial year - client 2, invoice 2")
	assert.Equal(suite.T(), "30 NET", results[2]["Payment terms"], "Payment terms - client 2, invoice 2")
	assert.Equal(suite.T(), "320.00", results[2]["Original amount"], "Original amount - client 2, invoice 2")
	assert.Equal(suite.T(), "320.00", results[2]["Outstanding amount"], "Outstanding amount - client 2, invoice 2")
	assert.Equal(suite.T(), "0", results[2]["Current"], "Current - client 2, invoice 2")
	assert.Equal(suite.T(), "0", results[2]["0-1 years"], "0-1 years - client 2, invoice 2")
	assert.Equal(suite.T(), "320.00", results[2]["1-2 years"], "1-2 years - client 2, invoice 2")
	assert.Equal(suite.T(), "0", results[2]["2-3 years"], "2-3 years - client 2, invoice 2")
	assert.Equal(suite.T(), "0", results[2]["3-5 years"], "3-5 years - client 2, invoice 2")
	assert.Equal(suite.T(), "0", results[2]["5+ years"], "5+ years - client 2, invoice 2")
	assert.Equal(suite.T(), "=\"3-5\"", results[2]["Debt impairment years"], "Debt impairment years - client 2, invoice 2")

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
}
