package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_cheque_payments_schedules() {
	ctx := suite.ctx
	today := suite.seeder.Today()
	oneMonthAgo := today.Sub(0, 1, 0)
	courtRef1 := "12345678"
	courtRef2 := "87654321"
	courtRef3 := "10101010"
	general := "320.00"

	// client 1
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", courtRef1, "1234")
	_, invoice1Reference := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeS2, &general, oneMonthAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 123, today.Date(), courtRef1, shared.TransactionTypeSupervisionChequePayment, today.Date(), 123456)
	suite.seeder.CreatePayment(ctx, 345, today.Date(), courtRef1, shared.TransactionTypeSupervisionChequePayment, today.Date(), 123456)

	// client 2
	client2ID := suite.seeder.CreateClient(ctx, "Alan", "Intelligence", courtRef2, "1234")
	_, invoice2Reference := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS2, &general, oneMonthAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 10000, today.Date(), courtRef2, shared.TransactionTypeSupervisionChequePayment, today.Date(), 123456)

	// client 3
	client3ID := suite.seeder.CreateClient(ctx, "C", "Lient", courtRef3, "1234")
	_, invoice3Reference := suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeAD, &general, oneMonthAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 5555, today.Date(), courtRef3, shared.TransactionTypeSupervisionChequePayment, today.Date(), 123456)

	c := Client{suite.seeder.Conn}

	rows, err := c.Run(ctx, &ChequePaymentsSchedule{
		Date:      &shared.Date{Time: today.Date()},
		PisNumber: 123456,
	})

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 5, len(rows))

	results := mapByHeader(rows)

	assert.NotEmpty(suite.T(), results)

	assert.Equal(suite.T(), courtRef1, results[0]["Court reference"], "Court ref - Client 1 cheque 1")
	assert.Equal(suite.T(), invoice1Reference, results[0]["Invoice reference"], "Invoice ref - Client 1 cheque 1")
	assert.Equal(suite.T(), "1.23", results[0]["Amount"], "Amount - Client 1 cheque 1")
	assert.Equal(suite.T(), today.Date().Format("2006-01-02"), results[0]["Payment date"], "Payment date - Client 1 cheque 1")
	assert.Equal(suite.T(), today.Date().Format("2006-01-02"), results[0]["Bank date"], "Bank date - Client 1 cheque 1")
	assert.Equal(suite.T(), today.Date().Format("2006-01-02"), results[0]["Create date"], "Create date - Client 1 cheque 1")

	assert.Equal(suite.T(), courtRef1, results[0]["Court reference"], "Court ref - Client 1 cheque 2")
	assert.Equal(suite.T(), invoice1Reference, results[1]["Invoice reference"], "Invoice ref - Client 1 cheque 2")
	assert.Equal(suite.T(), "3.45", results[1]["Amount"], "Amount - Client 1 cheque 2")

	assert.Equal(suite.T(), courtRef2, results[2]["Court reference"], "Court ref - Client 2 cheque 1")
	assert.Equal(suite.T(), invoice2Reference, results[2]["Invoice reference"], "Invoice ref - Client 2 cheque 1")
	assert.Equal(suite.T(), "100.00", results[2]["Amount"], "Amount - Client 2 cheque 1")
	assert.Equal(suite.T(), today.Date().Format("2006-01-02"), results[2]["Payment date"], "Payment date - Client 2 cheque 1")
	assert.Equal(suite.T(), today.Date().Format("2006-01-02"), results[2]["Bank date"], "Bank date - Client 2 cheque 1")
	assert.Equal(suite.T(), today.Date().Format("2006-01-02"), results[2]["Create date"], "Create date - Client 2 cheque 1")

	assert.Equal(suite.T(), courtRef3, results[3]["Court reference"], "Court ref - Client 3")
	assert.Equal(suite.T(), invoice3Reference, results[3]["Invoice reference"], "Invoice ref - Client 3")
	assert.Equal(suite.T(), "55.55", results[3]["Amount"], "Amount - Client 3")
	assert.Equal(suite.T(), today.Date().Format("2006-01-02"), results[3]["Payment date"], "Payment date - Client 3")
	assert.Equal(suite.T(), today.Date().Format("2006-01-02"), results[3]["Bank date"], "Bank date - Client 3")
	assert.Equal(suite.T(), today.Date().Format("2006-01-02"), results[3]["Create date"], "Create date - Client 3")
}
