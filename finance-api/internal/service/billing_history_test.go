package service

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func (suite *IntegrationSuite) TestService_GetBillingHistory() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (7, 1, '1234', 'DEMANDED', NULL);",
		"INSERT INTO finance_client VALUES (3, 2, '1234', 'DEMANDED', NULL);",
		"INSERT INTO invoice VALUES (1, 1, 7, 'S2', 'S203531/19', '2019-04-01', '2020-03-31', 32000, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, NULL, '2019-06-06', 99);",
		"INSERT INTO ledger VALUES (1, 'random1223', '2022-04-11T00:00:00+00:00', '', 12300, '', 'CREDIT MEMO', 'PENDING', 7, 1,NULL, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);",
		"INSERT INTO ledger VALUES (2, 'different', '2025-04-11T00:00:00+00:00', '', 55555, '', 'DEBIT MEMO', 'PENDING', 7, 2, NULL, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2025', 65);",
		"INSERT INTO ledger_allocation VALUES (1, 1, 1, '2022-04-11T00:00:00+00:00', 12300, 'PENDING', NULL, 'Notes here', '2022-04-11', NULL);",
		"INSERT INTO ledger_allocation VALUES (2, 2, 1, '2022-04-11T00:00:00+00:00', 55555, 'PENDING', NULL, 'Notes here', '2022-04-11', NULL);",
		"INSERT INTO fee_reduction VALUES (1, 7, 'REMISSION', NULL, '2020-03-31', '2023-03-31', 'Remission awarded', FALSE, '2020-03-03', '2020-04-01', 1);",
		"INSERT INTO fee_reduction VALUES (2, 7, 'HARDSHIP', NULL, '2019-04-01', '2020-03-31', 'Legacy (no created date) - do not display', FALSE, '2019-05-01');",
		"INSERT INTO fee_reduction VALUES (3, 7, 'REMISSION', NULL, '2020-03-31', '2023-03-31', 'Remission to see the notes', FALSE, '2020-03-03', '2020-04-01', 1, '2021-04-01', 2, 'Cancelled text here');",
		"INSERT INTO fee_reduction VALUES (4, 7, 'REMISSION', NULL, '2020-03-31', '2023-03-31', 'Remission approved', FALSE, '2020-03-03', '2020-04-01', 1);",
		"INSERT INTO ledger VALUES (3, 'different2', '2025-04-12T00:00:00+00:00', '', 12300, '', 'DEBIT MEMO', 'APPROVED', 7, 3, 4, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2025', 65);",
		"INSERT INTO ledger_allocation VALUES (3, 3, 1, '2022-04-11T00:00:00+00:00', 12300, 'PENDING', NULL, 'Notes here', '2022-04-11', NULL);",
	)

	Store := store.New(conn)

	debtMemoDate, _ := time.Parse("2006-01-02", "2025-04-11")
	creditMemoDate, _ := time.Parse("2006-01-02", "2022-04-11")
	invoiceDate, _ := time.Parse("2006-01-02", "2020-03-20")
	reductionStartDate, _ := time.Parse("2006-01-02", "2020-03-31")
	reductionEndDate, _ := time.Parse("2006-01-02", "2023-03-31")
	reductionReceivedDate, _ := time.Parse("2006-01-02", "2020-03-03")
	awardedReductionDate, _ := time.Parse("2006-01-02", "2020-04-01")
	appliedReductionDate, _ := time.Parse("2006-01-02", "2025-04-12")
	cancelledReductionDate, _ := time.Parse("2006-01-02", "2021-04-01")

	tests := []struct {
		name    string
		id      int
		want    []shared.BillingHistory
		wantErr bool
	}{
		{
			name: "returns all events that match the client id",
			id:   1,
			want: []shared.BillingHistory{
				{
					User: 1,
					Date: shared.Date{Time: appliedReductionDate},
					Event: shared.FeeReductionApplied{
						ClientId:      1,
						ReductionType: shared.FeeReductionTypeRemission,
						PaymentBreakdown: shared.PaymentBreakdown{
							InvoiceReference: shared.InvoiceEvent{
								ID:        1,
								Reference: "S203531/19",
							},
							Amount: 12300,
						},
						BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionApplied},
					},
					OutstandingBalance: 19700,
				},
				{
					User: 65,
					Date: shared.Date{Time: debtMemoDate},
					Event: shared.InvoiceAdjustmentPending{
						AdjustmentType: shared.AdjustmentTypeDebitMemo,
						ClientId:       1,
						Notes:          "",
						PaymentBreakdown: shared.PaymentBreakdown{
							InvoiceReference: shared.InvoiceEvent{
								ID:        1,
								Reference: "S203531/19",
							},
							Amount: 55555,
						},
						BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceAdjustmentPending},
					},
					OutstandingBalance: 32000,
				},
				{
					User: 65,
					Date: shared.Date{Time: creditMemoDate},
					Event: shared.InvoiceAdjustmentPending{
						AdjustmentType: shared.AdjustmentTypeCreditMemo,
						ClientId:       1,
						Notes:          "",
						PaymentBreakdown: shared.PaymentBreakdown{
							InvoiceReference: shared.InvoiceEvent{
								ID:        1,
								Reference: "S203531/19",
							},
							Amount: 12300,
						},
						BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceAdjustmentPending},
					},
					OutstandingBalance: 32000,
				},
				{
					User: 2,
					Date: shared.Date{Time: cancelledReductionDate},
					Event: shared.FeeReductionCancelled{
						ReductionType:      shared.FeeReductionTypeRemission,
						CancellationReason: "Cancelled text here",
						BaseBillingEvent:   shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionCancelled},
					},
					OutstandingBalance: 32000,
				},
				{
					User: 1,
					Date: shared.Date{Time: awardedReductionDate},
					Event: shared.FeeReductionAwarded{
						ReductionType:    shared.FeeReductionTypeRemission,
						StartDate:        shared.Date{Time: reductionStartDate},
						EndDate:          shared.Date{Time: reductionEndDate},
						DateReceived:     shared.Date{Time: reductionReceivedDate},
						Notes:            "Remission awarded",
						BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionAwarded},
					},
					OutstandingBalance: 32000,
				},
				{
					User: 1,
					Date: shared.Date{Time: awardedReductionDate},
					Event: shared.FeeReductionAwarded{
						ReductionType:    shared.FeeReductionTypeRemission,
						StartDate:        shared.Date{Time: reductionStartDate},
						EndDate:          shared.Date{Time: reductionEndDate},
						DateReceived:     shared.Date{Time: reductionReceivedDate},
						Notes:            "Remission approved",
						BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionAwarded},
					},
					OutstandingBalance: 32000,
				},
				{
					User: 99,
					Date: shared.Date{Time: invoiceDate},
					Event: shared.InvoiceGenerated{
						ClientId: 1,
						InvoiceReference: shared.InvoiceEvent{
							ID:        1,
							Reference: "S203531/19",
						},
						InvoiceType:      shared.InvoiceTypeS2,
						Amount:           32000,
						BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceGenerated},
					},
					OutstandingBalance: 32000,
				},
			},
		},
		{
			name: "returns an empty array when no match is found",
			id:   2,
			want: []shared.BillingHistory{},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			s := &Service{
				store: Store,
			}
			got, err := s.GetBillingHistory(suite.ctx, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBillingHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err == nil) && len(tt.want) == 0 {
				assert.Empty(t, got)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBillingHistory() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_computeBillingHistory(t *testing.T) {
	history := []historyHolder{
		{
			billingHistory: shared.BillingHistory{
				Date: shared.NewDate("2021-01-01"),
				Event: shared.InvoiceGenerated{
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceGenerated},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 3200,
		},
		{
			billingHistory: shared.BillingHistory{
				Date: shared.NewDate("2022-01-01"),
				Event: shared.InvoiceAdjustmentPending{
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceAdjustmentPending},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				Date: shared.NewDate("2025-01-01"),
				Event: shared.InvoiceGenerated{
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceGenerated},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 1000,
		},
		{
			billingHistory: shared.BillingHistory{
				Date: shared.NewDate("2024-01-01"),
				Event: shared.FeeReductionCancelled{
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionCancelled},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				Date: shared.NewDate("2023-01-01"),
				Event: shared.FeeReductionAwarded{
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionAwarded},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
	}

	expected := []shared.BillingHistory{
		{
			Date: shared.NewDate("2025-01-01"),
			Event: shared.InvoiceGenerated{
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceGenerated},
			},
			OutstandingBalance: 4200,
		},
		{
			Date: shared.NewDate("2024-01-01"),
			Event: shared.FeeReductionCancelled{
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionCancelled},
			},
			OutstandingBalance: 3200,
		},
		{
			Date: shared.NewDate("2023-01-01"),
			Event: shared.FeeReductionAwarded{
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionAwarded},
			},
			OutstandingBalance: 3200,
		},
		{
			Date: shared.NewDate("2022-01-01"),
			Event: shared.InvoiceAdjustmentPending{
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceAdjustmentPending},
			},
			OutstandingBalance: 3200,
		},
		{
			Date: shared.NewDate("2021-01-01"),
			Event: shared.InvoiceGenerated{
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceGenerated},
			},
			OutstandingBalance: 3200,
		},
	}

	assert.Equalf(t, expected, computeBillingHistory(history), "computeBillingHistory(%v)", history)
}

func Test_invoiceEvents(t *testing.T) {
	invoices := []store.GetGeneratedInvoicesRow{
		{
			InvoiceID:   1,
			Reference:   "AD123455/01",
			Feetype:     "AD",
			Amount:      100000,
			CreatedbyID: pgtype.Int4{Int32: 3, Valid: true},
			InvoiceDate: pgtype.Timestamp{Time: time.Date(2027, time.March, 31, 0, 0, 0, 0, time.UTC), Valid: true},
		},
	}

	expected := []historyHolder{{
		billingHistory: shared.BillingHistory{
			User: 3,
			Date: shared.Date{Time: time.Date(2027, time.March, 31, 0, 0, 0, 0, time.UTC)},
			Event: shared.InvoiceGenerated{
				ClientId: 1,
				InvoiceReference: shared.InvoiceEvent{
					ID:        1,
					Reference: "AD123455/01",
				},
				InvoiceType:      shared.InvoiceTypeAD,
				Amount:           100000,
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceGenerated},
			},
			OutstandingBalance: 0,
		},
		balanceAdjustment: 100000,
	}}

	assert.Equalf(t, expected, invoiceEvents(invoices, 1), "invoiceEvents(%v)", invoices)
}

func Test_processFeeReductionEvents(t *testing.T) {
	now := time.Now()
	reductions := []store.GetFeeReductionEventsRow{
		{
			Type:               "HARDSHIP",
			Startdate:          pgtype.Date{Time: now, Valid: true},
			Enddate:            pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true},
			Datereceived:       pgtype.Date{Time: now.Add(48 * time.Hour), Valid: true},
			Notes:              "Awarded",
			CreatedAt:          pgtype.Timestamp(pgtype.Date{Time: now.Add(72 * time.Hour), Valid: true}),
			CreatedBy:          pgtype.Int4{Int32: 1, Valid: true},
			CancelledAt:        pgtype.Timestamp{},
			CancelledBy:        pgtype.Int4{},
			CancellationReason: pgtype.Text{},
		},
		{
			Type:               "REMISSION",
			Startdate:          pgtype.Date{Time: now, Valid: true},
			Enddate:            pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true},
			Datereceived:       pgtype.Date{Time: now.Add(48 * time.Hour), Valid: true},
			Notes:              "Awarded",
			CreatedAt:          pgtype.Timestamp{Time: now.Add(72 * time.Hour), Valid: true},
			CreatedBy:          pgtype.Int4{Int32: 1, Valid: true},
			CancelledAt:        pgtype.Timestamp{Time: now.Add(96 * time.Hour), Valid: true},
			CancelledBy:        pgtype.Int4{Int32: 2, Valid: true},
			CancellationReason: pgtype.Text{String: "Cancelled for reasons", Valid: true},
		},
	}

	expected := []historyHolder{
		{
			billingHistory: shared.BillingHistory{
				User: 1,
				Date: shared.Date{Time: now.Add(72 * time.Hour)},
				Event: shared.FeeReductionAwarded{
					ReductionType:    shared.FeeReductionTypeHardship,
					StartDate:        shared.Date{Time: now},
					EndDate:          shared.Date{Time: now.Add(24 * time.Hour)},
					DateReceived:     shared.Date{Time: now.Add(48 * time.Hour)},
					Notes:            "Awarded",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionAwarded},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 2,
				Date: shared.Date{Time: now.Add(96 * time.Hour)},
				Event: shared.FeeReductionCancelled{
					ReductionType:      shared.FeeReductionTypeRemission,
					CancellationReason: "Cancelled for reasons",
					BaseBillingEvent:   shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionCancelled},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
	}

	assert.Equalf(t, expected, processFeeReductionEvents(reductions), "processFeeReductionEvents(%v)", reductions)
}
