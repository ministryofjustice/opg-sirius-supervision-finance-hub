package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_payments_schedules() {
	ctx := suite.ctx
	today := suite.seeder.Today()
	yesterday := suite.seeder.Today().Sub(0, 0, 1)
	oneMonthAgo := suite.seeder.Today().Sub(0, 1, 0)
	courtRef1 := "12345678"
	courtRef2 := "87654321"
	general := "320.00"

	// client 1
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", courtRef1, "1234")
	_, inv1Ref := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeS2, &general, oneMonthAgo.StringPtr(), nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 10000, yesterday.Date(), courtRef1, shared.TransactionTypeOPGBACSPayment, yesterday.Date())
	suite.seeder.CreatePayment(ctx, 11011, today.Date(), courtRef1, shared.TransactionTypeOPGBACSPayment, yesterday.Date())

	// client 2
	client2ID := suite.seeder.CreateClient(ctx, "Alan", "Intelligence", courtRef2, "1234")
	_, inv2Ref := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS2, &general, oneMonthAgo.StringPtr(), nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 12022, today.Date(), courtRef2, shared.TransactionTypeOPGBACSPayment, today.Date())
	suite.seeder.CreatePayment(ctx, 13033, today.Date(), courtRef2, shared.TransactionTypeMotoCardPayment, today.Date())

	c := Client{suite.seeder.Conn}

	tests := []struct {
		name         string
		date         shared.Date
		scheduleType shared.ReportScheduleType
		expectedRows int
		expectedData []map[string]string
	}{
		{
			name:         "filter by bank date",
			date:         shared.Date{Time: yesterday.Date()},
			scheduleType: shared.ReportOPGBACSTransfer,
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
			scheduleType: shared.ReportOPGBACSTransfer,
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
			scheduleType: shared.ReportTypeMOTOCardPayments,
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
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			rows, err := c.Run(ctx, &PaymentsSchedule{
				Date:         tt.date,
				ScheduleType: tt.scheduleType,
			})
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
