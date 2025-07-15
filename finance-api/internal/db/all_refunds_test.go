package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_all_refunds() {
	ctx := suite.ctx
	suite.seeder.CreateTestAssignee(suite.ctx)

	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)
	twoDaysAgo := today.Sub(0, 0, 2)
	threeDaysAgo := today.Sub(0, 0, 3)
	fourDaysAgo := today.Sub(0, 0, 4)
	aMonthAgo := today.Sub(0, 1, 0)

	// refund processed
	client1ID := suite.seeder.CreateClient(ctx, "Peter", "Processed", "11111111", "1234", "ACTIVE")
	suite.seeder.CreatePayment(ctx, 15000, fourDaysAgo.Date(), "11111111", shared.TransactionTypeMotoCardPayment, fourDaysAgo.Date(), 0)
	refund1ID := suite.seeder.CreateRefund(ctx, client1ID, "MR PETER PROCESSED", "11111110", "11-11-11", fourDaysAgo.Date())
	suite.seeder.SetRefundDecision(ctx, client1ID, refund1ID, shared.RefundStatusApproved, threeDaysAgo.Date())

	// refund fulfilled
	client6ID := suite.seeder.CreateClient(ctx, "Freddy", "Fulfilled", "66666666", "1234", "ACTIVE")
	suite.seeder.CreatePayment(ctx, 15000, yesterday.Date(), "66666666", shared.TransactionTypeMotoCardPayment, threeDaysAgo.Date(), 0)
	refund3ID := suite.seeder.CreateRefund(ctx, client6ID, "F FULFILLED", "66666660", "66-66-66", threeDaysAgo.Date())
	suite.seeder.SetRefundDecision(ctx, client6ID, refund3ID, shared.RefundStatusApproved, threeDaysAgo.Date())

	// refund cancelled
	client7ID := suite.seeder.CreateClient(ctx, "Conrad", "Cancelled", "77777777", "1234", "ACTIVE")
	suite.seeder.CreatePayment(ctx, 15000, yesterday.Date(), "77777777", shared.TransactionTypeMotoCardPayment, threeDaysAgo.Date(), 0)
	refund4ID := suite.seeder.CreateRefund(ctx, client7ID, "C CANCELLED", "77777770", "77-77-77", threeDaysAgo.Date())
	suite.seeder.SetRefundDecision(ctx, client7ID, refund4ID, shared.RefundStatusApproved, threeDaysAgo.Date())

	suite.seeder.ProcessApprovedRefunds(ctx, []int32{refund1ID, refund3ID, refund4ID}, twoDaysAgo.Date())
	suite.seeder.FulfillRefund(ctx, refund3ID, 15000, yesterday.Date(), "66666666", "F FULFILLED", "66666660", "66-66-66", yesterday.Date())

	suite.seeder.SetRefundDecision(ctx, client7ID, refund4ID, shared.RefundStatusCancelled, yesterday.Date())

	// refund pending
	client2ID := suite.seeder.CreateClient(ctx, "Percival", "Pending", "22222222", "1234", "ACTIVE")
	suite.seeder.CreatePayment(ctx, 15000, yesterday.Date(), "22222222", shared.TransactionTypeMotoCardPayment, yesterday.Date(), 0)
	_ = suite.seeder.CreateRefund(ctx, client2ID, "DR P PENDING", "22222220", "22-22-22", yesterday.Date())

	// refund approved
	client3ID := suite.seeder.CreateClient(ctx, "April", "Approved", "33333333", "1234", "ACTIVE")
	suite.seeder.CreatePayment(ctx, 15000, yesterday.Date(), "33333333", shared.TransactionTypeMotoCardPayment, yesterday.Date(), 0)
	refund2ID := suite.seeder.CreateRefund(ctx, client3ID, "MS APRIL APPROVED", "33333330", "33-33-33", yesterday.Date())
	suite.seeder.SetRefundDecision(ctx, client3ID, refund2ID, shared.RefundStatusApproved, today.Date())

	// too old
	client4ID := suite.seeder.CreateClient(ctx, "Oliver", "Old", "44444444", "1234", "ACTIVE")
	suite.seeder.CreatePayment(ctx, 15000, aMonthAgo.Date(), "44444444", shared.TransactionTypeMotoCardPayment, aMonthAgo.Date(), 0)
	_ = suite.seeder.CreateRefund(ctx, client4ID, "DR OLIVER OLD", "44444440", "44-44-44", aMonthAgo.Date())

	// too young
	client5ID := suite.seeder.CreateClient(ctx, "Yvonne", "Young", "55555555", "1234", "ACTIVE")
	suite.seeder.CreatePayment(ctx, 15000, today.Date(), "55555555", shared.TransactionTypeMotoCardPayment, today.Date(), 0)
	_ = suite.seeder.CreateRefund(ctx, client5ID, "PROF Y YOUNG", "55555550", "55-55-55", today.Date())

	c := Client{suite.seeder.Conn}
	from := shared.NewDate(fourDaysAgo.String())
	to := shared.NewDate(yesterday.String())

	rows, err := c.Run(ctx, NewAllRefunds(AllRefundsInput{
		FromDate: &from,
		ToDate:   &to,
	}))
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 6, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	assert.Equal(suite.T(), "22222222", results[0]["Court reference"], "Court reference - client 2")
	assert.Equal(suite.T(), "150.00", results[0]["Amount"], "Amount - client 2")
	assert.Equal(suite.T(), yesterday.String(), results[0]["Create date"], "Create date - client 2")
	assert.Equal(suite.T(), "Johnny Test", results[0]["Created by"], "Created by - client 2")
	assert.Equal(suite.T(), "", results[0]["Decision by"], "Decision by - client 2")
	assert.Equal(suite.T(), "PENDING", results[0]["Status"], "Status - client 2")
	assert.Equal(suite.T(), yesterday.String(), results[0]["Status Date"], "Status Date - client 2")

	assert.Equal(suite.T(), "33333333", results[1]["Court reference"], "Court reference - client 3")
	assert.Equal(suite.T(), "150.00", results[1]["Amount"], "Amount - client 3")
	assert.Equal(suite.T(), yesterday.String(), results[1]["Create date"], "Create date - client 3")
	assert.Equal(suite.T(), "Johnny Test", results[1]["Created by"], "Created by - client 3")
	assert.Equal(suite.T(), "Johnny Test", results[1]["Decision by"], "Decision by - client 3")
	assert.Equal(suite.T(), "APPROVED", results[1]["Status"], "Status - client 3")
	assert.Equal(suite.T(), today.String(), results[1]["Status Date"], "Status Date - client 3")

	assert.Equal(suite.T(), "66666666", results[2]["Court reference"], "Court reference - client 6")
	assert.Equal(suite.T(), "150.00", results[2]["Amount"], "Amount - client 6")
	assert.Equal(suite.T(), threeDaysAgo.String(), results[2]["Create date"], "Create date - client 6")
	assert.Equal(suite.T(), "Johnny Test", results[2]["Created by"], "Created by - client 6")
	assert.Equal(suite.T(), "Johnny Test", results[2]["Decision by"], "Decision by - client 6")
	assert.Equal(suite.T(), "FULFILLED", results[2]["Status"], "Status - client 6")
	assert.Equal(suite.T(), yesterday.String(), results[2]["Status Date"], "Status Date - client 6")

	assert.Equal(suite.T(), "77777777", results[3]["Court reference"], "Court reference - client 7")
	assert.Equal(suite.T(), "150.00", results[3]["Amount"], "Amount - client 7")
	assert.Equal(suite.T(), threeDaysAgo.String(), results[3]["Create date"], "Create date - client 7")
	assert.Equal(suite.T(), "Johnny Test", results[3]["Created by"], "Created by - client 7")
	assert.Equal(suite.T(), "Johnny Test", results[3]["Decision by"], "Decision by - client 7")
	assert.Equal(suite.T(), "CANCELLED", results[3]["Status"], "Status - client 7")
	assert.Equal(suite.T(), yesterday.String(), results[3]["Status Date"], "Status Date - client 7")

	assert.Equal(suite.T(), "11111111", results[4]["Court reference"], "Court reference - client 1")
	assert.Equal(suite.T(), "150.00", results[4]["Amount"], "Amount - client 1")
	assert.Equal(suite.T(), fourDaysAgo.String(), results[4]["Create date"], "Create date - client 1")
	assert.Equal(suite.T(), "Johnny Test", results[4]["Created by"], "Created by - client 1")
	assert.Equal(suite.T(), "Johnny Test", results[4]["Decision by"], "Decision by - client 1")
	assert.Equal(suite.T(), "PROCESSING", results[4]["Status"], "Status - client 1")
	assert.Equal(suite.T(), twoDaysAgo.String(), results[4]["Status Date"], "Status Date - client 1")
}
