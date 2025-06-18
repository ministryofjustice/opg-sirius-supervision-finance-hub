package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_invoices_schedules() {
	ctx := suite.ctx
	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)
	twoDaysAgo := today.Sub(0, 0, 2)
	oneMonthAgo := today.Sub(0, 1, 0)
	courtRef1 := "12345678"
	courtRef2 := "87654321"
	courtRef3 := "10101010"
	courtRef4 := "20202020"

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

	// client 4
	client4ID := suite.seeder.CreateClient(ctx, "Graham", "Simpson", courtRef4, "4321")
	_, adRef := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeAD, valToPtr("100.00"), twoDaysAgo.StringPtr(), nil, nil, nil, twoDaysAgo.StringPtr())
	_, s2Ref := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeS2, valToPtr("100.00"), twoDaysAgo.StringPtr(), nil, nil, nil, twoDaysAgo.StringPtr())
	_, s3Ref := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeS3, valToPtr("100.00"), twoDaysAgo.StringPtr(), nil, nil, nil, twoDaysAgo.StringPtr())
	_, b2Ref := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeB2, valToPtr("100.00"), twoDaysAgo.StringPtr(), nil, nil, nil, twoDaysAgo.StringPtr())
	_, b3Ref := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeB3, valToPtr("100.00"), twoDaysAgo.StringPtr(), nil, nil, nil, twoDaysAgo.StringPtr())
	_, sfGenRef := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeSF, valToPtr("100.00"), twoDaysAgo.StringPtr(), nil, nil, valToPtr("GENERAL"), twoDaysAgo.StringPtr())
	_, sfMinRef := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeSF, valToPtr("100.00"), twoDaysAgo.StringPtr(), nil, nil, valToPtr("MINIMAL"), twoDaysAgo.StringPtr())
	_, seGenRef := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeSE, valToPtr("100.00"), twoDaysAgo.StringPtr(), nil, nil, valToPtr("GENERAL"), twoDaysAgo.StringPtr())
	_, seMinRef := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeSE, valToPtr("100.00"), twoDaysAgo.StringPtr(), nil, nil, valToPtr("MINIMAL"), twoDaysAgo.StringPtr())
	_, soGenRef := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeSO, valToPtr("100.00"), twoDaysAgo.StringPtr(), nil, nil, valToPtr("GENERAL"), twoDaysAgo.StringPtr())
	_, soMinRef := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeSO, valToPtr("100.00"), twoDaysAgo.StringPtr(), nil, nil, valToPtr("MINIMAL"), twoDaysAgo.StringPtr())
	_, gaRef := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeGA, valToPtr("200.00"), twoDaysAgo.StringPtr(), nil, nil, nil, twoDaysAgo.StringPtr())
	_, gsRef := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeGS, valToPtr("100.00"), twoDaysAgo.StringPtr(), nil, nil, nil, twoDaysAgo.StringPtr())
	_, gtRef := suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeGT, valToPtr("100.00"), twoDaysAgo.StringPtr(), nil, nil, nil, twoDaysAgo.StringPtr())

	// ignored as raised date in scope but created date out of scope
	_, _ = suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeSE, valToPtr("10.00"), yesterday.StringPtr(), nil, nil, valToPtr("MINIMAL"), today.StringPtr())
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
		{
			name:         "ScheduleTypeAdFeeInvoices",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeAdFeeInvoices,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": adRef,
					"Amount":            "100.00",
					"Raised date":       twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "ScheduleTypeS2FeeInvoices",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeS2FeeInvoices,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": s2Ref,
					"Amount":            "100.00",
					"Raised date":       twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "ScheduleTypeS3FeeInvoices",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeS3FeeInvoices,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": s3Ref,
					"Amount":            "100.00",
					"Raised date":       twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "ScheduleTypeB2FeeInvoices",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeB2FeeInvoices,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": b2Ref,
					"Amount":            "100.00",
					"Raised date":       twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "ScheduleTypeB3FeeInvoices",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeB3FeeInvoices,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": b3Ref,
					"Amount":            "100.00",
					"Raised date":       twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "ScheduleTypeSFFeeInvoicesGeneral",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeSFFeeInvoicesGeneral,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": sfGenRef,
					"Amount":            "100.00",
					"Raised date":       twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "ScheduleTypeSFFeeInvoicesMinimal",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeSFFeeInvoicesMinimal,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": sfMinRef,
					"Amount":            "100.00",
					"Raised date":       twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "ScheduleTypeSEFeeInvoicesGeneral",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeSEFeeInvoicesGeneral,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": seGenRef,
					"Amount":            "100.00",
					"Raised date":       twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "ScheduleTypeSEFeeInvoicesMinimal",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeSEFeeInvoicesMinimal,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": seMinRef,
					"Amount":            "100.00",
					"Raised date":       twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "ScheduleTypeSOFeeInvoicesGeneral",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeSOFeeInvoicesGeneral,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": soGenRef,
					"Amount":            "100.00",
					"Raised date":       twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "ScheduleTypeSOFeeInvoicesMinimal",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeSOFeeInvoicesMinimal,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": soMinRef,
					"Amount":            "100.00",
					"Raised date":       twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "ScheduleTypeGAFeeInvoices",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGAFeeInvoices,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": gaRef,
					"Amount":            "200.00",
					"Raised date":       twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "ScheduleTypeGSFeeInvoices",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGSFeeInvoices,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": gsRef,
					"Amount":            "100.00",
					"Raised date":       twoDaysAgo.String(),
				},
			},
		},
		{
			name:         "ScheduleTypeGTFeeInvoices",
			date:         shared.Date{Time: twoDaysAgo.Date()},
			scheduleType: shared.ScheduleTypeGTFeeInvoices,
			expectedRows: 2,
			expectedData: []map[string]string{
				{
					"Court reference":   courtRef4,
					"Invoice reference": gtRef,
					"Amount":            "100.00",
					"Raised date":       twoDaysAgo.String(),
				},
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			rows, err := c.Run(ctx, NewInvoicesSchedule(InvoicesScheduleInput{
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
