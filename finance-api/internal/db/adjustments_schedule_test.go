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
	courtRef3 := "33333333"
	general := "320.00"

	// client 1
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", courtRef1, "1234")
	inv1Id, inv1Ref := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeS2, &general, twoYearsAgo.StringPtr(), nil, nil, valToPtr("GENERAL"))
	suite.seeder.CreateFeeReduction(ctx, client1ID, shared.FeeReductionTypeHardship, strconv.Itoa(twoYearsAgo.Sub(1, 0, 0).Date().Year()), 2, "notes", yesterday.Date())
	adjustment1Id := suite.seeder.CreateAdjustment(ctx, client1ID, inv1Id, shared.AdjustmentTypeCreditMemo, 30000, "Credit added") // unapplies should not add additional rows
	suite.seeder.ApproveAdjustment(ctx, client1ID, adjustment1Id)

	// client 2
	client2ID := suite.seeder.CreateClient(ctx, "Barry", "Giggle", courtRef2, "4321")
	suite.seeder.CreateFeeReduction(ctx, client2ID, shared.FeeReductionTypeRemission, strconv.Itoa(twoYearsAgo.Date().Year()), 3, "notes", twoYearsAgo.Date())
	_, inv2Ref := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeAD, nil, today.StringPtr(), nil, nil, nil)

	// client 3
	client3ID := suite.seeder.CreateClient(ctx, "Dani", "Debit", courtRef3, "4321")
	suite.seeder.CreateFeeReduction(ctx, client3ID, shared.FeeReductionTypeRemission, strconv.Itoa(twoYearsAgo.Date().Year()), 3, "notes", twoYearsAgo.Date()) // fee reduction to add credit that can be debited
	inv3Id, inv3Ref := suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeAD, nil, today.Sub(0, 6, 0).StringPtr(), nil, nil, nil)
	adjustment2Id := suite.seeder.CreateAdjustment(ctx, client3ID, inv3Id, shared.AdjustmentTypeDebitMemo, 5000, "Debit added")
	suite.seeder.ApproveAdjustment(ctx, client3ID, adjustment2Id)

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
		{
			name:         "debts",
			date:         shared.Date{Time: today.Date()},
			scheduleType: shared.ScheduleTypeADManualDebits,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef3,
					"Invoice reference": inv3Ref,
					"Amount":            "50.00",
					"Created date":      today.String(),
				},
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			rows, err := c.Run(ctx, &AdjustmentsSchedule{
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
