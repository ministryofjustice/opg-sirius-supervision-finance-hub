package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"strconv"
)

func (suite *IntegrationSuite) Test_credits_schedules() {
	ctx := suite.ctx
	today := suite.seeder.Today()
	yesterday := suite.seeder.Today().Sub(0, 0, 1)
	twoYearsAgo := suite.seeder.Today().Sub(2, 0, 0)
	courtRef1 := "12345678"
	courtRef2 := "87654321"
	general := "320.00"

	// client 1
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", courtRef1, "1234")
	inv1Id, inv1Ref := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeS2, &general, twoYearsAgo.StringPtr(), nil, nil, valToPtr("GENERAL"))
	suite.seeder.CreateFeeReduction(ctx, client1ID, shared.FeeReductionTypeHardship, strconv.Itoa(twoYearsAgo.Sub(1, 0, 0).Date().Year()), 2, "notes", yesterday.Date())
	adjustmentId := suite.seeder.CreateAdjustment(ctx, client1ID, inv1Id, shared.AdjustmentTypeCreditMemo, 30000, "Credit added") // unapplies should not add additional rows
	suite.seeder.ApproveAdjustment(ctx, client1ID, adjustmentId)

	// client 2
	client2ID := suite.seeder.CreateClient(ctx, "Barry", "Giggle", courtRef2, "4321")
	suite.seeder.CreateFeeReduction(ctx, client2ID, shared.FeeReductionTypeRemission, strconv.Itoa(twoYearsAgo.Date().Year()), 3, "notes", twoYearsAgo.Date())
	_, inv2Ref := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeAD, nil, today.StringPtr(), nil, nil, nil)

	c := Client{suite.seeder.Conn}

	tests := []struct {
		name         string
		date         shared.Date
		scheduleType shared.ScheduleType
		expectedRows int
		expectedData []map[string]string
	}{
		{
			name:         "filter by bank date",
			date:         shared.Date{Time: today.Date()},
			scheduleType: shared.ScheduleTypeADFeeReductions,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef2,
					"Invoice reference": inv2Ref,
					"Amount":            "50.00",
					"Created date":      today.String(),
				},
			},
		},
		{
			name:         "display ledgers not allocations (ignore unapplies)",
			date:         shared.Date{Time: today.Date()},
			scheduleType: shared.ScheduleTypeGeneralManualCredits,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef1,
					"Invoice reference": inv1Ref,
					"Amount":            "300.00",
					"Created date":      today.String(),
				},
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			rows, err := c.Run(ctx, &CreditsSchedule{
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
