package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_cheque_payments_schedules() {
	ctx := suite.ctx
	today := suite.seeder.Today()
	oneMonthAgo := suite.seeder.Today().Sub(0, 1, 0)
	courtRef1 := "12345678"
	courtRef2 := "87654321"
	courtRef3 := "10101010"
	general := "320.00"

	// client 1
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", courtRef1, "1234")
	_, invoice1Reference := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeS2, &general, oneMonthAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreateChequePayment(ctx, 10000, today.Date(), courtRef1, 123456, today.Date())
	suite.seeder.CreateChequePayment(ctx, 11011, today.Date(), courtRef1, 654321, today.Date())

	// client 2
	client2ID := suite.seeder.CreateClient(ctx, "Alan", "Intelligence", courtRef2, "1234")
	_, invoice2Reference := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS2, &general, oneMonthAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreateChequePayment(ctx, 12022, today.Date(), courtRef2, 123456, today.Date())
	suite.seeder.CreateChequePayment(ctx, 13033, today.Date(), courtRef2, 654321, today.Date())

	// client 3
	client3ID := suite.seeder.CreateClient(ctx, "C", "Lient", courtRef3, "1234")
	_, invoice3Reference := suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeAD, nil, oneMonthAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreateChequePayment(ctx, 8000, today.Date(), courtRef3, 123456, today.Date())

	c := Client{suite.seeder.Conn}

	rows, err := c.Run(ctx, &ChequePaymentsSchedule{
		Date:      &shared.Date{Time: today.Date()},
		PisNumber: 123456,
	})

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 4, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	assert.Equal(suite.T(), courtRef1, results[0]["Court reference"], "Court ref - Client 1")
	assert.Equal(suite.T(), invoice1Reference, results[0]["Invoice reference"], "Invoice ref - Client 1")
	assert.Equal(suite.T(), "100.00", results[0]["Amount"], "Amount - Client 1")
	assert.Equal(suite.T(), today.Date().Format("2006-01-02"), results[0]["Payment date"], "Payment date - Client 1")
	assert.Equal(suite.T(), today.Date().Format("2006-01-02"), results[0]["Bank date"], "Bank date - Client 1")
	assert.Equal(suite.T(), today.Date().Format("2006-01-02"), results[0]["Create date"], "Create date - Client 1")

	assert.Equal(suite.T(), courtRef2, results[1]["Court reference"], "Court ref - Client 2 cheque 1")
	assert.Equal(suite.T(), invoice2Reference, results[1]["Invoice reference"], "Invoice ref - Client 2 cheque 1")
	assert.Equal(suite.T(), "120.22", results[1]["Amount"], "Amount - Client 2 cheque 1")
	assert.Equal(suite.T(), today.Date().Format("2006-01-02"), results[1]["Payment date"], "Payment date - Client 2 cheque 1")
	assert.Equal(suite.T(), today.Date().Format("2006-01-02"), results[1]["Bank date"], "Bank date - Client 2 cheque 1")
	assert.Equal(suite.T(), today.Date().Format("2006-01-02"), results[1]["Create date"], "Create date - Client 2 cheque 1")

	assert.Equal(suite.T(), courtRef3, results[2]["Court reference"], "Court ref - Client 3")
	assert.Equal(suite.T(), invoice3Reference, results[2]["Invoice reference"], "Invoice ref - Client 3")
	assert.Equal(suite.T(), "80.00", results[2]["Amount"], "Amount - Client 3")
	assert.Equal(suite.T(), today.Date().Format("2006-01-02"), results[2]["Payment date"], "Payment date - Client 3")
	assert.Equal(suite.T(), today.Date().Format("2006-01-02"), results[2]["Bank date"], "Bank date - Client 3")
	assert.Equal(suite.T(), today.Date().Format("2006-01-02"), results[2]["Create date"], "Create date - Client 3")

}
