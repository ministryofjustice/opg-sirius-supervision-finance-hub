package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"strconv"
)

func (suite *IntegrationSuite) Test_adjustments_schedules() {
	ctx := suite.ctx
	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)
	twoDaysAgo := today.Sub(0, 0, 2)
	sixMonthsAgo := today.Sub(0, 6, 0)
	twoYearsAgo := today.Sub(2, 0, 0)
	courtRef1 := "12345678"
	courtRef2 := "87654321"
	courtRef3 := "33333333"
	courtRef4 := "44444444"
	general := "320.00"

	// client 1
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", courtRef1, "ACTIVE")
	inv1Id, inv1Ref := suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeS2, &general, twoYearsAgo.StringPtr(), nil, nil, valToPtr("GENERAL"), twoYearsAgo.StringPtr())
	_ = suite.seeder.CreateFeeReduction(ctx, client1ID, shared.FeeReductionTypeHardship, strconv.Itoa(twoYearsAgo.Sub(1, 0, 0).Date().Year()), 2, "notes", yesterday.Date())
	suite.seeder.CreateAdjustment(ctx, client1ID, inv1Id, shared.AdjustmentTypeCreditMemo, 30000, "Credit added", nil) // unapplies should not add additional rows

	// client 2
	client2ID := suite.seeder.CreateClient(ctx, "Barry", "Giggle", courtRef2, "ACTIVE")
	_ = suite.seeder.CreateFeeReduction(ctx, client2ID, shared.FeeReductionTypeRemission, strconv.Itoa(twoYearsAgo.Date().Year()), 3, "notes", twoYearsAgo.Date())
	_, inv2Ref := suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeAD, nil, today.StringPtr(), nil, nil, nil, nil)

	// client 3
	client3ID := suite.seeder.CreateClient(ctx, "Dani", "Debit", courtRef3, "ACTIVE")
	_ = suite.seeder.CreateFeeReduction(ctx, client3ID, shared.FeeReductionTypeRemission, strconv.Itoa(twoYearsAgo.Date().Year()), 3, "notes", twoYearsAgo.Date()) // fee reduction to add credit that can be debited
	inv3Id, inv3Ref := suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeAD, nil, sixMonthsAgo.StringPtr(), nil, nil, nil, sixMonthsAgo.StringPtr())
	suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeS2, &general, sixMonthsAgo.StringPtr(), nil, nil, valToPtr("GENERAL"), nil) // ignore as not AD
	suite.seeder.CreateAdjustment(ctx, client3ID, inv3Id, shared.AdjustmentTypeDebitMemo, 4500, "Debit added", nil)

	// client 4 - fee reductions
	client4ID := suite.seeder.CreateClient(ctx, "Alison", "Adjustments", courtRef4, "ACTIVE")
	// create fee reduction
	_ = suite.seeder.CreateFeeReduction(ctx, client4ID, shared.FeeReductionTypeHardship, strconv.Itoa(twoYearsAgo.Date().Year()), 3, "notes", twoYearsAgo.Date())
	// create one of each invoice
	invADID, invADRef := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeAD, valToPtr("100.00"), twoDaysAgo.StringPtr(), nil, nil, nil, twoDaysAgo.StringPtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invADID, shared.AdjustmentTypeDebitMemo, 1000, "Debit added", twoDaysAgo.DatePtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invADID, shared.AdjustmentTypeWriteOff, 0, "Write off", twoDaysAgo.DatePtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invADID, shared.AdjustmentTypeCreditMemo, 1000, "Credit added", twoDaysAgo.DatePtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invADID, shared.AdjustmentTypeWriteOffReversal, 0, "Write off reversal", twoDaysAgo.DatePtr())

	invS2ID, invS2Ref := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeS2, valToPtr("320.00"), twoDaysAgo.StringPtr(), nil, nil, nil, twoDaysAgo.StringPtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invS2ID, shared.AdjustmentTypeDebitMemo, 1000, "Debit added", twoDaysAgo.DatePtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invS2ID, shared.AdjustmentTypeWriteOff, 0, "Write off", twoDaysAgo.DatePtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invS2ID, shared.AdjustmentTypeCreditMemo, 1000, "Credit added", twoDaysAgo.DatePtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invS2ID, shared.AdjustmentTypeWriteOffReversal, 0, "Write off reversal", twoDaysAgo.DatePtr())

	invS3ID, invS3Ref := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeS3, valToPtr("10.00"), twoDaysAgo.StringPtr(), nil, nil, nil, twoDaysAgo.StringPtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invS3ID, shared.AdjustmentTypeDebitMemo, 1000, "Debit added", twoDaysAgo.DatePtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invS3ID, shared.AdjustmentTypeWriteOff, 0, "Write off", twoDaysAgo.DatePtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invS3ID, shared.AdjustmentTypeCreditMemo, 1000, "Credit added", twoDaysAgo.DatePtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invS3ID, shared.AdjustmentTypeWriteOffReversal, 0, "Write off reversal", twoDaysAgo.DatePtr())

	invGAID, invGARef := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeGA, valToPtr("200.00"), twoDaysAgo.StringPtr(), nil, nil, nil, twoDaysAgo.StringPtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invGAID, shared.AdjustmentTypeDebitMemo, 1000, "Debit added", twoDaysAgo.DatePtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invGAID, shared.AdjustmentTypeWriteOff, 0, "Write off", twoDaysAgo.DatePtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invGAID, shared.AdjustmentTypeCreditMemo, 1000, "Credit added", twoDaysAgo.DatePtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invGAID, shared.AdjustmentTypeWriteOffReversal, 0, "Write off reversal", twoDaysAgo.DatePtr())

	invGSID, invGSRef := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeGS, valToPtr("100.00"), twoDaysAgo.StringPtr(), nil, nil, nil, twoDaysAgo.StringPtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invGSID, shared.AdjustmentTypeDebitMemo, 1000, "Debit added", twoDaysAgo.DatePtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invGSID, shared.AdjustmentTypeWriteOff, 0, "Write off", twoDaysAgo.DatePtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invGSID, shared.AdjustmentTypeCreditMemo, 1000, "Credit added", twoDaysAgo.DatePtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invGSID, shared.AdjustmentTypeWriteOffReversal, 0, "Write off reversal", twoDaysAgo.DatePtr())

	invGTID, invGTRef := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeGT, valToPtr("100.00"), twoDaysAgo.StringPtr(), nil, nil, nil, twoDaysAgo.StringPtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invGTID, shared.AdjustmentTypeDebitMemo, 1000, "Debit added", twoDaysAgo.DatePtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invGTID, shared.AdjustmentTypeWriteOff, 0, "Write off", twoDaysAgo.DatePtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invGTID, shared.AdjustmentTypeCreditMemo, 1000, "Credit added", twoDaysAgo.DatePtr())
	suite.seeder.CreateAdjustment(ctx, client4ID, invGTID, shared.AdjustmentTypeWriteOffReversal, 0, "Write off reversal", twoDaysAgo.DatePtr())
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
					"Amount":            "45.00",
					"Created date":      today.String(),
				},
			},
		},
		{
			name:         "filter by AD",
			date:         shared.Date{Time: sixMonthsAgo.Date()},
			scheduleType: shared.ScheduleTypeADFeeReductions,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef3,
					"Invoice reference": inv3Ref,
					"Amount":            "50.00",
					"Created date":      sixMonthsAgo.String(),
				},
			},
		},
		// AD invoice adjustments
		{
			name:         "AD invoice fee reduction",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeADFeeReductions,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invADRef,
					"Amount":            "100.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "credit for AD invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeADManualCredits,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invADRef,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "debit for AD invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeADManualDebits,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invADRef,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "write off for AD invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeADWriteOffs,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invADRef,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "write off reversal for AD invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeADWriteOffReversals,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invADRef,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		// general invoice adjustments
		{
			name:         "S2 invoice fee reduction",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGeneralFeeReductions,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invS2Ref,
					"Amount":            "320.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "credit for general invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGeneralManualCredits,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invS2Ref,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "debit for general invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGeneralManualDebits,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invS2Ref,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "write off for general invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGeneralWriteOffs,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invS2Ref,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "write off reversal for general invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGeneralWriteOffReversals,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invS2Ref,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		// minimal invoice adjustments
		{
			name:         "S3 invoice fee reduction",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeMinimalFeeReductions,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invS3Ref,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "credit for minimal invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeMinimalManualCredits,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invS3Ref,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "debit for minimal invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeMinimalManualDebits,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invS3Ref,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "write off for minimal invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeMinimalWriteOffs,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invS3Ref,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "write off reversal for minimal invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeMinimalWriteOffReversals,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invS3Ref,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		// GA invoice adjustments
		{
			name:         "GA invoice fee reduction",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGAFeeReductions,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invGARef,
					"Amount":            "200.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "credit for GA invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGAManualCredits,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invGARef,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "debit for GA invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGAManualDebits,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invGARef,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "write off for GA invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGAWriteOffs,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invGARef,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "write off reversal for GA invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGAWriteOffReversals,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invGARef,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		// GS invoice adjustments
		{
			name:         "GS invoice fee reduction",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGSFeeReductions,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invGSRef,
					"Amount":            "100.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "credit for GS invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGSManualCredits,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invGSRef,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "debit for GS invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGSManualDebits,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invGSRef,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "write off for GS invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGSWriteOffs,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invGSRef,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "write off reversal for GS invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGSWriteOffReversals,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invGSRef,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		// GT invoice adjustments
		{
			name:         "GT invoice fee reduction",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGTFeeReductions,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invGTRef,
					"Amount":            "100.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "credit for GT invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGTManualCredits,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invGTRef,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "debit for GT invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGTManualDebits,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invGTRef,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "write off for GT invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGTWriteOffs,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invGTRef,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "write off reversal for GT invoice",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGTWriteOffReversals,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": invGTRef,
					"Amount":            "10.00",
					"Created date":      twoDaysAgo.String(),
				},
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			rows, err := c.Run(ctx, NewAdjustmentsSchedule(AdjustmentsScheduleInput{
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
