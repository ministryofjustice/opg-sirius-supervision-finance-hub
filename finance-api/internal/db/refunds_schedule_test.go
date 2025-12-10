package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_refunds_schedules() {
	ctx := suite.ctx
	suite.seeder.CreateTestAssignee(suite.ctx)

	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)
	twoDaysAgo := today.Sub(0, 0, 2)
	threeDaysAgo := today.Sub(0, 0, 3)
	fourDaysAgo := today.Sub(0, 0, 4)
	fiveDaysAgo := today.Sub(0, 0, 5)
	sixDaysAgo := today.Sub(0, 0, 6)
	sevenDaysAgo := today.Sub(0, 0, 7)
	eightDaysAgo := today.Sub(0, 0, 8)

	// refund processed
	client1ID := suite.seeder.CreateClient(ctx, "Filly", "Fulfilled", "11111111", "1234", "ACTIVE")
	suite.seeder.CreatePayment(ctx, 15000, fourDaysAgo.Date(), "11111111", shared.TransactionTypeMotoCardPayment, fourDaysAgo.Date(), 0)
	refund1ID := suite.seeder.CreateRefund(ctx, client1ID, "MRS F FULFILLED", "11111110", "11-11-11", fourDaysAgo.Date())
	suite.seeder.SetRefundDecision(ctx, client1ID, refund1ID, shared.RefundStatusApproved, threeDaysAgo.Date())

	// refund fulfilled
	client2ID := suite.seeder.CreateClient(ctx, "Freddy", "Fulfilled", "22222222", "1234", "ACTIVE")
	suite.seeder.CreatePayment(ctx, 13050, yesterday.Date(), "22222222", shared.TransactionTypeMotoCardPayment, threeDaysAgo.Date(), 0)
	refund2ID := suite.seeder.CreateRefund(ctx, client2ID, "MR F FULFILLED", "22222220", "22-22-22", threeDaysAgo.Date())
	suite.seeder.SetRefundDecision(ctx, client2ID, refund2ID, shared.RefundStatusApproved, threeDaysAgo.Date())

	// two days ago
	suite.seeder.ProcessApprovedRefunds(ctx, []int32{refund1ID, refund2ID}, twoDaysAgo.Date())
	suite.seeder.FulfillRefund(ctx, refund1ID, 15000, twoDaysAgo.Date(), "11111111", "MRS F FULFILLED", "11111110", "111111", twoDaysAgo.Date())
	suite.seeder.FulfillRefund(ctx, refund2ID, 13050, twoDaysAgo.Date(), "22222222", "MR F FULFILLED", "22222220", "222222", yesterday.Date())

	// one day ago
	client3ID := suite.seeder.CreateClient(ctx, "Frederick", "Fulfilled Jr", "33333333", "", "1234")
	suite.seeder.CreatePayment(ctx, 12345, yesterday.Date(), "33333333", shared.TransactionTypeMotoCardPayment, threeDaysAgo.Date(), 0)
	refund3ID := suite.seeder.CreateRefund(ctx, client3ID, "MR FULFILLED JR", "33333330", "33-33-33", threeDaysAgo.Date())
	suite.seeder.SetRefundDecision(ctx, client3ID, refund3ID, shared.RefundStatusApproved, threeDaysAgo.Date())

	suite.seeder.ProcessApprovedRefunds(ctx, []int32{refund3ID}, twoDaysAgo.Date())
	suite.seeder.FulfillRefund(ctx, refund1ID, 12345, yesterday.Date(), "33333333", "MRS F FULFILLED", "33333330", "333333", yesterday.Date())

	// reversed refund
	client4ID := suite.seeder.CreateClient(ctx, "Randy", "Reversed", "44444444", "1234", "ACTIVE")
	suite.seeder.CreatePayment(ctx, 14000, today.Date(), "44444444", shared.TransactionTypeMotoCardPayment, fiveDaysAgo.Date(), 0)
	refund4ID := suite.seeder.CreateRefund(ctx, client4ID, "MR R REVERSED", "44444440", "44-44-44", fiveDaysAgo.Date())
	suite.seeder.SetRefundDecision(ctx, client4ID, refund4ID, shared.RefundStatusApproved, fiveDaysAgo.Date())

	suite.seeder.ProcessApprovedRefunds(ctx, []int32{refund4ID}, fiveDaysAgo.Date())
	suite.seeder.FulfillRefund(ctx, refund4ID, 14000, fiveDaysAgo.Date(), "44444444", "MR R REVERSED", "44444440", "444444", fiveDaysAgo.Date())
	suite.seeder.ReverseRefund(ctx, "44444444", "140.00", fiveDaysAgo.Date(), fiveDaysAgo.Date())

	// reversed refund that allocates to invoice created in between refund fulfilled and reversal
	client5ID := suite.seeder.CreateClient(ctx, "Ivy", "Invoice", "55555555", "1234", "ACTIVE")
	suite.seeder.CreatePayment(ctx, 16000, sevenDaysAgo.Date(), "55555555", shared.TransactionTypeMotoCardPayment, sevenDaysAgo.Date(), 0)
	refund5ID := suite.seeder.CreateRefund(ctx, client5ID, "MS I INVOICE", "55555550", "55-55-55", sevenDaysAgo.Date())
	suite.seeder.SetRefundDecision(ctx, client5ID, refund5ID, shared.RefundStatusApproved, sevenDaysAgo.Date())

	suite.seeder.ProcessApprovedRefunds(ctx, []int32{refund5ID}, sevenDaysAgo.Date())
	suite.seeder.FulfillRefund(ctx, refund5ID, 16000, sevenDaysAgo.Date(), "55555555", "MS I INVOICE", "55555550", "555555", sevenDaysAgo.Date())

	_, _ = suite.seeder.CreateInvoice(ctx, client5ID, shared.InvoiceTypeAD, nil, sixDaysAgo.StringPtr(), nil, nil, nil, sixDaysAgo.StringPtr())
	suite.seeder.ReverseRefund(ctx, "55555555", "160.00", sevenDaysAgo.Date(), sixDaysAgo.Date())

	// as above but all occurs on the same day
	client6ID := suite.seeder.CreateClient(ctx, "Sally", "SameDay", "66666666", "1234", "ACTIVE")
	suite.seeder.CreatePayment(ctx, 16000, eightDaysAgo.Date(), "66666666", shared.TransactionTypeMotoCardPayment, eightDaysAgo.Date(), 0)
	refund6ID := suite.seeder.CreateRefund(ctx, client6ID, "MS S SAMEDAY", "66666660", "66-66-66", eightDaysAgo.Date())
	suite.seeder.SetRefundDecision(ctx, client6ID, refund6ID, shared.RefundStatusApproved, eightDaysAgo.Date())

	suite.seeder.ProcessApprovedRefunds(ctx, []int32{refund6ID}, eightDaysAgo.Date())
	suite.seeder.FulfillRefund(ctx, refund6ID, 16000, eightDaysAgo.Date(), "66666666", "MS S SAMEDAY", "66666660", "666666", eightDaysAgo.Date())
	_, _ = suite.seeder.CreateInvoice(ctx, client6ID, shared.InvoiceTypeAD, nil, eightDaysAgo.StringPtr(), nil, nil, nil, eightDaysAgo.StringPtr())
	suite.seeder.ReverseRefund(ctx, "66666666", "160.00", eightDaysAgo.Date(), eightDaysAgo.Date())

	c := Client{suite.seeder.Conn}

	tests := []struct {
		name         string
		date         shared.Date
		scheduleType shared.ScheduleType
		expectedRows int
		expectedData []map[string]string
	}{
		{
			name:         "refunds",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			expectedRows: 3,
			expectedData: []map[string]string{
				{
					"Court reference":         "11111111",
					"Amount":                  "150.00",
					"Bank date":               twoDaysAgo.String(),
					"Fulfilled (create) date": twoDaysAgo.String(),
				},
				{
					"Court reference":         "22222222",
					"Amount":                  "130.50",
					"Bank date":               twoDaysAgo.String(),
					"Fulfilled (create) date": yesterday.String(),
				},
			},
		},
		{
			name:         "filtered by date",
			date:         shared.Date{Time: yesterday.Date()},
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":         "33333333",
					"Amount":                  "123.45",
					"Bank date":               yesterday.String(),
					"Fulfilled (create) date": yesterday.String(),
				},
			},
		},
		{
			name:         "reversed refund on same day",
			date:         shared.Date{Time: fiveDaysAgo.Date()},
			expectedRows: 3,
			expectedData: []map[string]string{
				{
					"Court reference":         "44444444",
					"Amount":                  "140.00",
					"Bank date":               fiveDaysAgo.String(),
					"Fulfilled (create) date": fiveDaysAgo.String(),
				},
				{
					"Court reference":         "44444444",
					"Amount":                  "-140.00",
					"Bank date":               fiveDaysAgo.String(),
					"Fulfilled (create) date": fiveDaysAgo.String(),
				},
			},
		},
		{
			name:         "reversed refund different day, partially allocated to invoice",
			date:         shared.Date{Time: sixDaysAgo.Date()},
			expectedRows: 3,
			expectedData: []map[string]string{
				{
					"Court reference":         "55555555",
					"Amount":                  "-100.00",
					"Bank date":               sixDaysAgo.String(),
					"Fulfilled (create) date": sixDaysAgo.String(),
				},
				{
					"Court reference":         "55555555",
					"Amount":                  "-60.00",
					"Bank date":               sixDaysAgo.String(),
					"Fulfilled (create) date": sixDaysAgo.String(),
				},
			},
		},
		{
			name:         "reversed refund same day, partially allocated to invoice",
			date:         shared.Date{Time: eightDaysAgo.Date()},
			expectedRows: 4,
			expectedData: []map[string]string{
				{
					"Court reference":         "66666666",
					"Amount":                  "160.00",
					"Bank date":               eightDaysAgo.String(),
					"Fulfilled (create) date": eightDaysAgo.String(),
				},
				{
					"Court reference":         "66666666",
					"Amount":                  "-100.00",
					"Bank date":               eightDaysAgo.String(),
					"Fulfilled (create) date": eightDaysAgo.String(),
				},
				{
					"Court reference":         "66666666",
					"Amount":                  "-60.00",
					"Bank date":               eightDaysAgo.String(),
					"Fulfilled (create) date": eightDaysAgo.String(),
				},
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			rows, err := c.Run(ctx, NewRefundsSchedule(RefundsScheduleInput{
				Date: &tt.date,
			}))

			assert.NoError(suite.T(), err)
			assert.Equal(suite.T(), tt.expectedRows, len(rows))

			results := mapByHeader(rows)
			assert.NotEmpty(suite.T(), results)

			for i, expected := range tt.expectedData {
				for key, value := range expected {
					assert.Equal(suite.T(), value, results[i][key], tt.name+": "+key)
				}
			}
		})
	}
}
