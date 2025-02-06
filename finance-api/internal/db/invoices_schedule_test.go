package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_invoices_schedules() {
	ctx := suite.ctx
	yesterday := suite.seeder.Today().Sub(0, 0, 1)
	oneMonthAgo := suite.seeder.Today().Sub(0, 1, 0)
	courtRef1 := "12345678"
	courtRef2 := "87654321"
	courtRef3 := "10101010"

	// client 1
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", courtRef1, "1234")
	_, inv1Ref := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, valToPtr("100.00"), oneMonthAgo.StringPtr(), nil, nil, nil, oneMonthAgo.StringPtr())
	_, inv2Ref := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeSE, valToPtr("320.00"), yesterday.StringPtr(), nil, nil, valToPtr("GENERAL"), yesterday.StringPtr())

	// client 2
	client2ID := suite.seeder.CreateClient(ctx, "Alan", "Intelligence", courtRef2, "1234")
	_, inv3Ref := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeSE, valToPtr("300.88"), yesterday.StringPtr(), nil, nil, valToPtr("GENERAL"), yesterday.StringPtr())

	// client 3
	client3ID := suite.seeder.CreateClient(ctx, "Barry", "Giggle", courtRef3, "4321")
	_, inv4Ref := suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeSE, valToPtr("10.00"), yesterday.StringPtr(), nil, nil, valToPtr("MINIMAL"), yesterday.StringPtr())

	// ignored as raised date in scope but created date out of scope
	_, _ = suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeSE, valToPtr("10.00"), yesterday.StringPtr(), nil, nil, valToPtr("MINIMAL"), suite.seeder.Today().StringPtr())
	_, _ = suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeSE, valToPtr("10.00"), oneMonthAgo.StringPtr(), nil, nil, valToPtr("MINIMAL"), oneMonthAgo.StringPtr())

	c := Client{suite.seeder.Conn}

	tests := []struct {
		name         string
		date         shared.Date
		scheduleType shared.ScheduleType
		expectedRows int
		expectedData []map[string]string
	}{
		{
			name:         "filter by invoice date",
			date:         shared.Date{Time: oneMonthAgo.Date()},
			scheduleType: shared.ScheduleTypeAdFeeInvoices,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef1,
					"Invoice reference": inv1Ref,
					"Amount":            "100.00",
					"Raised date":       oneMonthAgo.String(),
				},
			},
		},
		{
			name:         "multi client filter by invoice type",
			date:         shared.Date{Time: yesterday.Date()},
			scheduleType: shared.ScheduleTypeSEFeeInvoicesGeneral,
			expectedRows: 3,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef1,
					"Invoice reference": inv2Ref,
					"Amount":            "320.00",
					"Raised date":       yesterday.String(),
				},
				{
					"Court reference":   courtRef2,
					"Invoice reference": inv3Ref,
					"Amount":            "300.88",
					"Raised date":       yesterday.String(),
				},
			},
		},
		{
			name:         "filter by supervision level",
			date:         shared.Date{Time: yesterday.Date()},
			scheduleType: shared.ScheduleTypeSEFeeInvoicesMinimal,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef3,
					"Invoice reference": inv4Ref,
					"Amount":            "10.00",
					"Raised date":       yesterday.String(),
				},
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			rows, err := c.Run(ctx, &InvoicesSchedule{
				Date:         &tt.date,
				ScheduleType: &tt.scheduleType,
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
