package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_approved_refunds() {
	ctx := suite.ctx
	suite.seeder.CreateTestAssignee(suite.ctx)

	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)

	// refund already processed
	client1ID := suite.seeder.CreateClient(ctx, "Peter", "Processed", "11111111", "1234")
	suite.seeder.CreatePayment(ctx, 15000, yesterday.Date(), "11111111", shared.TransactionTypeMotoCardPayment, today.Date(), 0)
	refund1ID := suite.seeder.CreateRefund(ctx, client1ID, "MR PETER PROCESSED", "11111110", "11-11-11")
	suite.seeder.SetRefundDecision(ctx, client1ID, refund1ID, shared.RefundStatusApproved)
	suite.seeder.ProcessApprovedRefunds(ctx)

	// refund pending
	client2ID := suite.seeder.CreateClient(ctx, "Percival", "Pending", "22222222", "1234")
	suite.seeder.CreatePayment(ctx, 15000, yesterday.Date(), "22222222", shared.TransactionTypeMotoCardPayment, today.Date(), 0)
	_ = suite.seeder.CreateRefund(ctx, client2ID, "DR P PENDING", "22222220", "22-22-22")

	// refund approved
	client3ID := suite.seeder.CreateClient(ctx, "April", "Approved", "33333333", "1234")
	suite.seeder.CreatePayment(ctx, 15000, yesterday.Date(), "33333333", shared.TransactionTypeMotoCardPayment, today.Date(), 0)
	refund2ID := suite.seeder.CreateRefund(ctx, client3ID, "MS APRIL APPROVED", "33333330", "33-33-33")
	suite.seeder.SetRefundDecision(ctx, client3ID, refund2ID, shared.RefundStatusApproved)

	c := Client{suite.seeder.Conn}

	rows, err := c.Run(ctx, &ApprovedRefunds{})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	assert.Equal(suite.T(), "33333333", results[0]["Court reference"], "Court reference - client 1")
	assert.Equal(suite.T(), "150.00", results[0]["Amount"], "Amount - client 1")
	assert.Equal(suite.T(), "MS APRIL APPROVED", results[0]["Bank account name"], "Bank account name - client 1")
	assert.Equal(suite.T(), "33333330", results[0]["Bank account number"], "Bank account number - client 1")
	assert.Equal(suite.T(), "=\"33-33-33\"", results[0]["Bank account sort code"], "Bank account sort code - client 1")
	assert.Equal(suite.T(), "Johnny Test", results[0]["Created by"], "Do Created by - client 1")
	assert.Equal(suite.T(), "Johnny Test", results[0]["Approved by"], "Approved by - client 1")
}
