package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"strconv"
)

func (suite *IntegrationSuite) Test_unapply_reapply_schedules() {
	ctx := suite.ctx

	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)
	sixMonthsAgo := today.Sub(0, 6, 0)
	twoYearsAgo := today.Sub(2, 0, 0)
	threeYearsAgo := today.Sub(3, 0, 0)
	courtRef1 := "12345678"
	courtRef2 := "87654321"
	courtRef3 := "33333333"
	general := "320.00"
	minimal := "10.00"

	// client 1 - no credit
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", courtRef1, "1234")
	_ = suite.seeder.CreateFeeReduction(ctx, client1ID, shared.FeeReductionTypeExemption, strconv.Itoa(twoYearsAgo.Date().Year()), 3, "notes", yesterday.Date())
	_, _ = suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeS2, &general, sixMonthsAgo.StringPtr(), nil, nil, valToPtr("GENERAL"), sixMonthsAgo.StringPtr())

	// client 2 - unapplied credit
	client2ID := suite.seeder.CreateClient(ctx, "Una", "Unapply", courtRef2, "4321")
	_ = suite.seeder.CreateFeeReduction(ctx, client2ID, shared.FeeReductionTypeExemption, strconv.Itoa(twoYearsAgo.Date().Year()), 3, "notes", twoYearsAgo.Date())
	inv2ID, inv2Ref := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeAD, nil, sixMonthsAgo.StringPtr(), nil, nil, nil, sixMonthsAgo.StringPtr())
	suite.seeder.CreateAdjustment(ctx, client2ID, inv2ID, shared.AdjustmentTypeCreditMemo, 9900, "Credit added", yesterday.DatePtr())

	// client 3 - reapplied credit
	client3ID := suite.seeder.CreateClient(ctx, "Reginald", "Reapply", courtRef3, "4321")
	_ = suite.seeder.CreateFeeReduction(ctx, client3ID, shared.FeeReductionTypeExemption, strconv.Itoa(threeYearsAgo.Date().Year()), 2, "notes", threeYearsAgo.Date())
	inv3ID, inv3Ref := suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeAD, nil, twoYearsAgo.StringPtr(), nil, nil, nil, twoYearsAgo.StringPtr())
	suite.seeder.CreateAdjustment(ctx, client3ID, inv3ID, shared.AdjustmentTypeCreditMemo, 8800, "Credit added", sixMonthsAgo.DatePtr())
	_, inv4Ref := suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeS3, &minimal, yesterday.StringPtr(), nil, nil, valToPtr("MINIMAL"), yesterday.StringPtr())

	c := Client{suite.seeder.Conn}

	tests := []struct {
		name         string
		date         shared.Date
		scheduleType shared.ScheduleType
		expectedRows int
		expectedData []map[string]string
	}{
		{
			name:         "unapplied payments",
			date:         shared.Date{Time: yesterday.Date()},
			scheduleType: shared.ScheduleTypeUnappliedPayments,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef2,
					"Invoice reference": inv2Ref,
					"Amount":            "99.00",
					"Created date":      yesterday.String(),
				},
			},
		},
		{
			name:         "reapplied payments",
			date:         shared.Date{Time: yesterday.Date()},
			scheduleType: shared.ScheduleTypeReappliedPayments,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef3,
					"Invoice reference": inv4Ref,
					"Amount":            "10.00",
					"Created date":      yesterday.String(),
				},
			},
		},
		{
			name:         "filtered by date",
			date:         shared.Date{Time: sixMonthsAgo.Date()},
			scheduleType: shared.ScheduleTypeUnappliedPayments,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef3,
					"Invoice reference": inv3Ref,
					"Amount":            "88.00",
					"Created date":      sixMonthsAgo.String(),
				},
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			rows, err := c.Run(ctx, NewUnapplyReapplySchedule(UnapplyReapplyScheduleInput{
				Date:         &tt.date,
				ScheduleType: &tt.scheduleType,
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
