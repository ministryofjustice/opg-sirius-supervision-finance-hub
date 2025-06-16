package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_payments_schedules() {
	ctx := suite.ctx
	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)
	oneMonthAgo := today.Sub(0, 1, 0)
	courtRef1 := "12345678"
	courtRef2 := "87654321"
	courtRef3 := "10101010"
	courtRef4 := "44444444"
	courtRef5 := "55555555"
	courtRef6 := "66666666"
	general := "320.00"

	// client 1
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", courtRef1, "1234")
	_, inv1Ref := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeS2, &general, oneMonthAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 10000, yesterday.Date(), courtRef1, shared.TransactionTypeOPGBACSPayment, yesterday.Date(), 0)
	suite.seeder.CreatePayment(ctx, 11011, today.Date(), courtRef1, shared.TransactionTypeOPGBACSPayment, yesterday.Date(), 0)

	// client 2
	client2ID := suite.seeder.CreateClient(ctx, "Alan", "Intelligence", courtRef2, "1234")
	_, inv2Ref := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS2, &general, oneMonthAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 12022, today.Date(), courtRef2, shared.TransactionTypeOPGBACSPayment, today.Date(), 0)
	suite.seeder.CreatePayment(ctx, 13033, today.Date(), courtRef2, shared.TransactionTypeMotoCardPayment, today.Date(), 0)

	// client 3
	client3ID := suite.seeder.CreateClient(ctx, "C", "Lient", courtRef3, "1234")
	_, inv3Ref := suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeAD, nil, oneMonthAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 12000, today.Date(), courtRef3, shared.TransactionTypeDirectDebitPayment, today.Date(), 0)

	// an online card payment that is misapplied and added onto the correct client
	client5ID := suite.seeder.CreateClient(ctx, "Ernie", "Error", courtRef4, "2314")
	_, inv4Ref := suite.seeder.CreateInvoice(ctx, client5ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())
	suite.seeder.CreatePayment(ctx, 15000, yesterday.Date(), courtRef4, shared.TransactionTypeOnlineCardPayment, yesterday.Date(), 0)

	client6ID := suite.seeder.CreateClient(ctx, "Colette", "Correct", courtRef5, "2314")
	_, inv5Ref := suite.seeder.CreateInvoice(ctx, client6ID, shared.InvoiceTypeS2, &general, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())
	suite.seeder.ReversePayment(ctx, courtRef4, courtRef5, "150.00", yesterday.Date(), yesterday.Date(), shared.TransactionTypeOnlineCardPayment, yesterday.Date())

	// cheques
	client7ID := suite.seeder.CreateClient(ctx, "Ian", "Test", courtRef6, "1234")
	_, inv6Ref := suite.seeder.CreateInvoice(ctx, client7ID, shared.InvoiceTypeS2, &general, oneMonthAgo.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 1000, today.Date(), courtRef6, shared.TransactionTypeSupervisionChequePayment, today.Date(), 123456)
	suite.seeder.CreatePayment(ctx, 1234, today.Date(), courtRef6, shared.TransactionTypeSupervisionChequePayment, today.Date(), 654321)

	c := Client{suite.seeder.Conn}

	tests := []struct {
		name         string
		date         shared.Date
		scheduleType shared.ScheduleType
		pisNumber    int
		expectedRows int
		expectedData []map[string]string
	}{
		{
			name:         "filter by bank date",
			date:         shared.Date{Time: yesterday.Date()},
			scheduleType: shared.ScheduleTypeOPGBACSTransfer,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef1,
					"Invoice reference": inv1Ref,
					"Amount":            "100.00",
					"Payment date":      yesterday.String(),
					"Bank date":         yesterday.String(),
				},
			},
		},
		{
			name:         "multi client filter by bank date",
			date:         shared.Date{Time: today.Date()},
			scheduleType: shared.ScheduleTypeOPGBACSTransfer,
			expectedRows: 3,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef1,
					"Invoice reference": inv1Ref,
					"Amount":            "110.11",
					"Payment date":      yesterday.String(),
					"Bank date":         today.String(),
				},
				{
					"Court reference":   courtRef2,
					"Invoice reference": inv2Ref,
					"Amount":            "120.22",
					"Payment date":      today.String(),
					"Bank date":         today.String(),
				},
			},
		},
		{
			name:         "filter by payment type",
			date:         shared.Date{Time: today.Date()},
			scheduleType: shared.ScheduleTypeMOTOCardPayments,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef2,
					"Invoice reference": inv2Ref,
					"Amount":            "130.33",
					"Payment date":      today.String(),
					"Bank date":         today.String(),
				},
			},
		},
		{
			name:         "overpayments",
			date:         shared.Date{Time: today.Date()},
			scheduleType: shared.ScheduleTypeDirectDebitPayments,
			expectedRows: 3,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef3,
					"Invoice reference": inv3Ref,
					"Amount":            "100.00",
					"Payment date":      today.String(),
					"Bank date":         today.String(),
				},
				{
					"Court reference":   courtRef3,
					"Invoice reference": "",
					"Amount":            "20.00",
					"Payment date":      today.String(),
					"Bank date":         today.String(),
				},
			},
		},
		{
			name:         "cheques by pis number",
			date:         shared.Date{Time: today.Date()},
			scheduleType: shared.ScheduleTypeChequePayments,
			pisNumber:    123456,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef6,
					"Invoice reference": inv6Ref,
					"Amount":            "10.00",
					"Payment date":      today.String(),
					"Bank date":         today.String(),
				},
			},
		},
		{
			name:         "misapplied payments with overpayment",
			date:         shared.Date{Time: yesterday.Date()},
			scheduleType: shared.ScheduleTypeOnlineCardPayments,
			expectedRows: 6,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": inv4Ref,
					"Amount":            "100.00",
					"Payment date":      yesterday.String(),
					"Bank date":         yesterday.String(),
				},
				{
					"Court reference":   courtRef4,
					"Invoice reference": "",
					"Amount":            "50.00",
					"Payment date":      yesterday.String(),
					"Bank date":         yesterday.String(),
				},
				{
					"Court reference":   courtRef4,
					"Invoice reference": inv4Ref,
					"Amount":            "-100.00",
					"Payment date":      yesterday.String(),
					"Bank date":         yesterday.String(),
				},
				{
					"Court reference":   courtRef4,
					"Invoice reference": "",
					"Amount":            "-50.00",
					"Payment date":      yesterday.String(),
					"Bank date":         yesterday.String(),
				},
				{ // this one is missing
					"Court reference":   courtRef5,
					"Invoice reference": inv5Ref,
					"Amount":            "150.00",
					"Payment date":      yesterday.String(),
					"Bank date":         yesterday.String(),
				},
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			rows, err := c.Run(ctx, NewPaymentsSchedule(&tt.date, &tt.scheduleType, tt.pisNumber))

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
