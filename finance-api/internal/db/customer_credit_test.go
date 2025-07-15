package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"strconv"
)

func (suite *IntegrationSuite) Test_customer_credit() {
	ctx := suite.ctx

	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)
	twoMonthsAgo := today.Sub(0, 2, 0)
	twoYearsAgo := today.Sub(2, 0, 0)
	threeYearsAgo := today.Sub(3, 0, 0)
	minimal := "10"

	// client 1 with:
	// - Credit balance due to overpayment
	// £100 - £223.45 = -£123.45
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12345678", "1234", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client1ID)
	suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 22345, twoYearsAgo.Date(), "12345678", shared.TransactionTypeOPGBACSPayment, today.Date(), 0)

	// client 2 with:
	// - Credit balance due to fee reduction
	// - Partially reapplied
	// £100 - £100 + £100 - £10 = -£90
	client2ID := suite.seeder.CreateClient(ctx, "John", "Suite", "87654321", "4321", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client2ID)
	suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeAD, nil, twoYearsAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 10000, twoYearsAgo.Date(), "87654321", shared.TransactionTypeOPGBACSPayment, twoYearsAgo.Date(), 0)
	_ = suite.seeder.CreateFeeReduction(ctx, client2ID, shared.FeeReductionTypeExemption, strconv.Itoa(threeYearsAgo.Date().Year()), 2, "A reduction", threeYearsAgo.Date())
	suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS3, &minimal, today.StringPtr(), today.StringPtr(), nil, nil, nil)

	// Doesn't display client with:
	// - No credit balance after unapplied funds fully reapplied
	// £100 - £150 + £100 = £50 (outstanding)
	client3ID := suite.seeder.CreateClient(ctx, "Billy", "Client", "23456789", "2345", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client3ID)
	suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 15000, yesterday.Date(), "23456789", shared.TransactionTypeOPGBACSPayment, yesterday.Date(), 0)
	suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeAD, nil, today.StringPtr(), nil, nil, nil, nil)

	c := Client{suite.seeder.Conn}

	rows, err := c.Run(ctx, NewCustomerCredit())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// client 1
	assert.Equal(suite.T(), "Ian Test", results[0]["Customer name"], "Customer name - client 1")
	assert.Equal(suite.T(), "12345678", results[0]["Customer number"], "Customer number - client 1")
	assert.Equal(suite.T(), "1234", results[0]["SOP number"], "SOP number - client 1")
	assert.Equal(suite.T(), "123.45", results[0]["Credit balance"], "Credit balance - client 1")

	// client 2
	assert.Equal(suite.T(), "John Suite", results[1]["Customer name"], "Customer name - client 2")
	assert.Equal(suite.T(), "87654321", results[1]["Customer number"], "Customer number - client 2")
	assert.Equal(suite.T(), "4321", results[1]["SOP number"], "SOP number - client 2")
	assert.Equal(suite.T(), "90.00", results[1]["Credit balance"], "Credit balance - client 2")
}
