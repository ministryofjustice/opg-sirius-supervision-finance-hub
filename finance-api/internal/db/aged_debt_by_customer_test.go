package db

import (
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"strconv"
)

func (suite *IntegrationSuite) Test_aged_debt_by_customer() {
	ctx := suite.ctx
	today := suite.seeder.Today()
	twoMonthsAgo := today.Sub(0, 2, 0)
	twoYearsAgo := today.Sub(2, 0, 0)
	fourYearsAgo := today.Sub(4, 0, 0)
	fiveYearsAgo := today.Sub(5, 0, 0)
	general := "320.00"
	// one client with:
	// - a lay deputy
	// - an active order
	// - one written off invoice
	// - one active invoice (today)
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12345678", "1234")
	suite.seeder.CreateDeputy(ctx, client1ID, "Suzie", "Deputy", "LAY")
	suite.seeder.CreateOrder(ctx, client1ID, "ACTIVE")
	unpaidInvoiceID, _ := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, today.StringPtr(), nil, nil, nil, nil)
	paidInvoiceID, _ := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, today.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreateAdjustment(ctx, client1ID, paidInvoiceID, shared.AdjustmentTypeWriteOff, 0, "Written off", nil)

	suite.seeder.SeedData(
		fmt.Sprintf("INSERT INTO supervision_finance.ledger VALUES (99, 'ignore-me', '2022-04-11T08:36:40+00:00', '', 99999, '', 'CREDIT REMISSION', 'APPROVED', %d, NULL, NULL, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 2);", client1ID),
		fmt.Sprintf("INSERT INTO supervision_finance.ledger_allocation VALUES (99, 99, %d, '2022-04-11T08:36:40+00:00', 99999, 'ALLOCATED', NULL, 'Notes here', '2022-04-11', NULL);", unpaidInvoiceID),
	)

	// one client with:
	// - a pro deputy
	// - a closed order
	// - one active invoice (2020) with hardship reduction
	// - one active invoice (2022)
	client2ID := suite.seeder.CreateClient(ctx, "John", "Suite", "87654321", "4321")
	suite.seeder.CreateDeputy(ctx, client2ID, "Jane", "Deputy", "PRO")
	suite.seeder.CreateOrder(ctx, client2ID, "CLOSED")
	_ = suite.seeder.CreateFeeReduction(ctx, client2ID, shared.FeeReductionTypeRemission, strconv.Itoa(fiveYearsAgo.Date().Year()), 2, "A reduction", fiveYearsAgo.Date())
	suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeAD, nil, fourYearsAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS2, &general, twoYearsAgo.StringPtr(), twoYearsAgo.StringPtr(), nil, nil, nil)

	// one client with:
	// - a PA deputy
	// - an active order
	// - one active invoice (two months old)
	client3ID := suite.seeder.CreateClient(ctx, "Billy", "Client", "23456789", "2345")
	suite.seeder.CreateDeputy(ctx, client3ID, "Local", "Authority", "PA")
	suite.seeder.CreateOrder(ctx, client3ID, "ACTIVE")
	suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, nil)

	c := Client{suite.seeder.Conn}

	rows, err := c.Run(ctx, NewAgedDebtByCustomer())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 4, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// client 1
	assert.Equal(suite.T(), "Ian Test", results[0]["Customer name"], "Customer name - client 1")
	assert.Equal(suite.T(), "12345678", results[0]["Customer number"], "Customer number - client 1")
	assert.Equal(suite.T(), "1234", results[0]["SOP number"], "SOP number - client 1")
	assert.Equal(suite.T(), "LAY", results[0]["Deputy type"], "Deputy type - client 1")
	assert.Equal(suite.T(), "Yes", results[0]["Active case?"], "Active case? - client 1")
	assert.Equal(suite.T(), "100.00", results[0]["Outstanding amount"], "Outstanding amount - client 1")
	assert.Equal(suite.T(), "100.00", results[0]["Current"], "Current - client 1")
	assert.Equal(suite.T(), "0", results[0]["1 - 21 Days"], "1 - 21 Days - client 1")
	assert.Equal(suite.T(), "0", results[0]["22 - 35 Days"], "22 - 35 Days - client 1")
	assert.Equal(suite.T(), "0", results[0]["36 - 65 Days"], "36 - 65 Days - client 1")
	assert.Equal(suite.T(), "0", results[0]["66 - 90 Days"], "66 - 90 Days - client 1")
	assert.Equal(suite.T(), "0", results[0]["91 - 120 Days"], "91 - 120 Days - client 1")
	assert.Equal(suite.T(), "0", results[0]["121 - 365 Days"], "121 - 365 Days - client 1")
	assert.Equal(suite.T(), "0", results[0]["0-1 years"], "0-1 years - client 1") // current "debt" is not yet debt
	assert.Equal(suite.T(), "0", results[0]["1-2 years"], "1-2 years - client 1")
	assert.Equal(suite.T(), "0", results[0]["2-3 years"], "2-3 years - client 1")
	assert.Equal(suite.T(), "0", results[0]["3-5 years"], "3-5 years - client 1")
	assert.Equal(suite.T(), "0", results[0]["5+ years"], "5+ years - client 1")

	// client 2
	assert.Equal(suite.T(), "John Suite", results[1]["Customer name"], "Customer name - client 2")
	assert.Equal(suite.T(), "87654321", results[1]["Customer number"], "Customer number - client 2")
	assert.Equal(suite.T(), "4321", results[1]["SOP number"], "SOP number - client 2")
	assert.Equal(suite.T(), "PRO", results[1]["Deputy type"], "Deputy type - client 2")
	assert.Equal(suite.T(), "No", results[1]["Active case?"], "Active case? - client 2")
	assert.Equal(suite.T(), "370.00", results[1]["Outstanding amount"], "Outstanding amount - client 2")
	assert.Equal(suite.T(), "0", results[1]["Current"], "Current - client 2")
	assert.Equal(suite.T(), "0", results[1]["1 - 21 Days"], "1 - 21 Days - client 2")
	assert.Equal(suite.T(), "0", results[1]["22 - 35 Days"], "22 - 35 Days - client 2")
	assert.Equal(suite.T(), "0", results[1]["36 - 65 Days"], "36 - 65 Days - client 2")
	assert.Equal(suite.T(), "0", results[1]["66 - 90 Days"], "66 - 90 Days - client 2")
	assert.Equal(suite.T(), "0", results[1]["91 - 120 Days"], "91 - 120 Days - client 2")
	assert.Equal(suite.T(), "0", results[1]["121 - 365 Days"], "121 - 365 Days - client 2")
	assert.Equal(suite.T(), "0", results[1]["0-1 years"], "0-1 years - client 2")
	assert.Equal(suite.T(), "0", results[1]["1-2 years"], "1-2 years - client 2")
	assert.Equal(suite.T(), "0", results[1]["2-3 years"], "2-3 years - client 2")
	assert.Equal(suite.T(), "370.00", results[1]["3-5 years"], "3-5 years - client 2")
	assert.Equal(suite.T(), "0", results[1]["5+ years"], "5+ years - client 2")

	// client 3
	assert.Equal(suite.T(), "Billy Client", results[2]["Customer name"], "Customer name - client 2")
	assert.Equal(suite.T(), "23456789", results[2]["Customer number"], "Customer number - client 2")
	assert.Equal(suite.T(), "2345", results[2]["SOP number"], "SOP number - client 2")
	assert.Equal(suite.T(), "PA", results[2]["Deputy type"], "Deputy type - client 2")
	assert.Equal(suite.T(), "Yes", results[2]["Active case?"], "Active case? - client 2")
	assert.Equal(suite.T(), "100.00", results[2]["Outstanding amount"], "Outstanding amount - client 2")
	assert.Equal(suite.T(), "0", results[2]["Current"], "Current - client 2")
	assert.Equal(suite.T(), "0", results[2]["1 - 21 Days"], "1 - 21 Days - client 2")
	assert.Equal(suite.T(), "100.00", results[2]["22 - 35 Days"], "22 - 35 Days - client 2")
	assert.Equal(suite.T(), "0", results[2]["36 - 65 Days"], "36 - 65 Days - client 2")
	assert.Equal(suite.T(), "0", results[2]["66 - 90 Days"], "66 - 90 Days - client 2")
	assert.Equal(suite.T(), "0", results[2]["91 - 120 Days"], "91 - 120 Days - client 2")
	assert.Equal(suite.T(), "0", results[2]["121 - 365 Days"], "121 - 365 Days - client 2")
	assert.Equal(suite.T(), "100.00", results[2]["0-1 years"], "0-1 years - client 2")
	assert.Equal(suite.T(), "0", results[2]["1-2 years"], "1-2 years - client 2")
	assert.Equal(suite.T(), "0", results[2]["2-3 years"], "2-3 years - client 2")
	assert.Equal(suite.T(), "0", results[2]["3-5 years"], "3-5 years - client 2")
	assert.Equal(suite.T(), "0", results[2]["5+ years"], "5+ years - client 2")
}
