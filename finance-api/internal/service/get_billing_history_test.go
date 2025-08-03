package service

import (
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func (suite *IntegrationSuite) TestService_GetBillingHistory() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1,1,1234,'DEMANDED',NULL);",
		"INSERT INTO invoice VALUES (9,1,1,'AD','AD000001/24','2024-10-07','2024-10-07',10000,NULL,'2024-10-07',NULL,'2024-10-07','Created manually',NULL,NULL,'2024-10-07 09:31:44',1);",
		"INSERT INTO invoice VALUES (10,1,1,'AD','AD000002/24','2024-10-07','2024-10-07',10000,NULL,'2024-10-07',NULL,'2024-10-07','Created manually',NULL,NULL,'2024-10-07 09:35:03',1);",
		"INSERT INTO invoice_adjustment VALUES (4,1,9,'2024-10-07','CREDIT WRITE OFF',10000,'Writing off','REJECTED','2024-10-07 09:32:23',1,'2024-10-07 09:33:24',1)",
		"INSERT INTO invoice_adjustment VALUES (5,1,9,'2024-10-07','CREDIT MEMO',10000,'Adding credit','APPROVED','2024-10-07 09:34:38',1,'2024-10-07 09:34:44',1)",
		"INSERT INTO fee_reduction VALUES (1, 1, 'HARDSHIP', NULL, '2019-04-01', '2020-03-31', 'Legacy (no created BankDate) - do not display', FALSE, '2019-05-01');",
		"INSERT INTO fee_reduction VALUES (5,1,'REMISSION',NULL,'2024-04-01','2027-03-31','Needs remission',TRUE,'2024-10-07','2024-10-07 09:32:50',1,'2024-10-07 09:33:19',1,'Wrong remission');",
		"INSERT INTO ledger VALUES (5,'09799ea2-5f8f-4ecb-8200-f021ab96def1','2024-10-07 09:32:50','',5000,'Credit due to approved remission','CREDIT REMISSION','CONFIRMED',1,NULL,5,NULL,NULL,NULL,NULL,NULL,NULL,NULL,1);",
		"INSERT INTO ledger VALUES (6,'6e469827-fff7-4c22-a2e2-8b7d3580350c','2024-10-07 09:34:44','',5000,'Credit due to approved credit memo','CREDIT MEMO','CONFIRMED',1,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,1);",
		"INSERT INTO ledger VALUES (7,'babda0f7-2f07-4b85-a991-7d45be9474e2','2024-10-07 09:35:03','',5000,'Excess credit applied to invoice','CREDIT REAPPLY','CONFIRMED',1,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,1);",
		"INSERT INTO ledger VALUES (8,'13b3851c-2e7d-43d0-86ad-86ffca586f57','2024-10-07 09:36:05','',1000,'Moto payment','MOTO CARD PAYMENT','CONFIRMED',1,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,1);",
		"INSERT INTO ledger_allocation VALUES (5,5,9,'2024-10-07 09:32:50',5000,'ALLOCATED');",
		"INSERT INTO ledger_allocation VALUES (6,6,9,'2024-10-07 09:34:44',10000,'ALLOCATED');",
		"INSERT INTO ledger_allocation VALUES (7,6,9,'2024-10-07 09:34:44',-5000,'UNAPPLIED',NULL,'Unapplied funds as a result of applying credit memo');",
		"INSERT INTO ledger_allocation VALUES (8,7,10,'2024-10-07 09:35:03',5000,'REAPPLIED');",
		"INSERT INTO ledger_allocation VALUES (9,8,10,'2024-10-07 09:36:05',1000,'ALLOCATED');",
	)

	Store := store.New(seeder.Conn)

	tests := []struct {
		name    string
		id      int32
		want    []shared.BillingHistory
		wantErr bool
	}{
		{
			name: "returns all events that match the client id",
			id:   1,
			want: []shared.BillingHistory{
				{
					User: 1,
					Date: shared.NewDate("2024-10-07 09:36:05"),
					Event: shared.PaymentProcessed{
						TransactionEvent: shared.TransactionEvent{
							ClientId:        1,
							TransactionType: shared.TransactionTypeMotoCardPayment,
							Amount:          1000,
							Breakdown: []shared.PaymentBreakdown{
								{
									InvoiceReference: shared.InvoiceEvent{
										ID:        10,
										Reference: "AD000002/24",
									},
									Amount: 1000,
									Status: "ALLOCATED",
								},
							},
							BaseBillingEvent: shared.BaseBillingEvent{
								Type: shared.EventTypePaymentProcessed,
							},
						},
					},
					OutstandingBalance: 4000,
					CreditBalance:      0,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-10-07 09:35:03"),
					Event: shared.FeeReductionApplied{
						TransactionEvent: shared.TransactionEvent{
							ClientId:        1,
							TransactionType: shared.TransactionTypeReapply,
							Amount:          5000,
							Breakdown: []shared.PaymentBreakdown{
								{
									InvoiceReference: shared.InvoiceEvent{
										ID:        10,
										Reference: "AD000002/24",
									},
									Amount: 5000,
									Status: "REAPPLIED",
								},
							},
							BaseBillingEvent: shared.BaseBillingEvent{
								Type: shared.EventTypeReappliedCredit,
							},
						},
					},
					OutstandingBalance: 5000,
					CreditBalance:      0,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-10-07 09:35:03"),
					Event: shared.InvoiceGenerated{
						ClientId: 1,
						InvoiceReference: shared.InvoiceEvent{
							ID:        10,
							Reference: "AD000002/24",
						},
						InvoiceType: shared.InvoiceTypeAD,
						Amount:      10000,
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeInvoiceGenerated,
						},
					},
					OutstandingBalance: 10000,
					CreditBalance:      5000,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-10-07 09:34:38"),
					Event: shared.InvoiceAdjustmentApplied{
						TransactionEvent: shared.TransactionEvent{
							ClientId:        1,
							TransactionType: shared.TransactionTypeCreditMemo,
							Amount:          10000,
							Breakdown: []shared.PaymentBreakdown{
								{
									InvoiceReference: shared.InvoiceEvent{
										ID:        9,
										Reference: "AD000001/24",
									},
									Amount: 10000,
									Status: "ALLOCATED",
								},
								{
									InvoiceReference: shared.InvoiceEvent{
										ID:        9,
										Reference: "AD000001/24",
									},
									Amount: 5000,
									Status: "UNAPPLIED",
								},
							},
							BaseBillingEvent: shared.BaseBillingEvent{
								Type: shared.EventTypeInvoiceAdjustmentApplied,
							},
						},
					},
					OutstandingBalance: 0,
					CreditBalance:      5000,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-10-07 09:34:38"),
					Event: shared.InvoiceAdjustmentPending{
						AdjustmentType: shared.AdjustmentTypeCreditMemo,
						ClientId:       1,
						Notes:          "Adding credit",
						PaymentBreakdown: shared.PaymentBreakdown{
							InvoiceReference: shared.InvoiceEvent{
								ID:        9,
								Reference: "AD000001/24",
							},
							Amount: 10000,
						},
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeInvoiceAdjustmentPending,
						},
					},
					OutstandingBalance: 5000,
					CreditBalance:      0,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-10-07 09:33:19"),
					Event: shared.FeeReductionCancelled{
						ReductionType:      shared.FeeReductionTypeRemission,
						CancellationReason: "Wrong remission",
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeFeeReductionCancelled,
						},
					},
					OutstandingBalance: 5000,
					CreditBalance:      0,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-10-07 09:32:50"),
					Event: shared.FeeReductionApplied{
						TransactionEvent: shared.TransactionEvent{
							ClientId:        1,
							TransactionType: shared.TransactionTypeRemission,
							Amount:          5000,
							Breakdown: []shared.PaymentBreakdown{
								{
									InvoiceReference: shared.InvoiceEvent{
										ID:        9,
										Reference: "AD000001/24",
									},
									Amount: 5000,
									Status: "ALLOCATED",
								},
							},
							BaseBillingEvent: shared.BaseBillingEvent{
								Type: shared.EventTypeFeeReductionApplied,
							},
						},
					},
					OutstandingBalance: 5000,
					CreditBalance:      0,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-10-07 09:32:50"),
					Event: shared.FeeReductionAwarded{
						ReductionType: shared.FeeReductionTypeRemission,
						StartDate:     shared.NewDate("2024-10-07"),
						EndDate:       shared.NewDate("2024-04-01"),
						DateReceived:  shared.NewDate("2027-03-31"),
						Notes:         "Needs remission",
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeFeeReductionAwarded,
						},
					},
					OutstandingBalance: 10000,
					CreditBalance:      0,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-10-07 09:32:23"),
					Event: shared.InvoiceAdjustmentPending{
						AdjustmentType: shared.AdjustmentTypeWriteOff,
						ClientId:       1,
						Notes:          "Writing off",
						PaymentBreakdown: shared.PaymentBreakdown{
							InvoiceReference: shared.InvoiceEvent{
								ID:        9,
								Reference: "AD000001/24",
							},
							Amount: 10000,
						},
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeInvoiceAdjustmentPending,
						},
					},
					OutstandingBalance: 10000,
					CreditBalance:      0,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-10-07 09:32:23"),
					Event: shared.InvoiceAdjustmentRejected{
						AdjustmentType: shared.AdjustmentTypeWriteOff,
						ClientId:       1,
						Notes:          "Writing off",
						PaymentBreakdown: shared.PaymentBreakdown{
							InvoiceReference: shared.InvoiceEvent{
								ID:        9,
								Reference: "AD000001/24",
							},
							Amount: 10000,
						},
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeInvoiceAdjustmentRejected,
						},
					},
					OutstandingBalance: 10000,
					CreditBalance:      0,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-10-07 09:31:44"),
					Event: shared.InvoiceGenerated{
						ClientId: 1,
						InvoiceReference: shared.InvoiceEvent{
							ID:        9,
							Reference: "AD000001/24",
						},
						InvoiceType: shared.InvoiceTypeAD,
						Amount:      10000,
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeInvoiceGenerated,
						},
					},
					OutstandingBalance: 10000,
					CreditBalance:      0,
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

			// need to compare the unmarshalled states due to using shared.TransactionEvent as the abstract type preventing DeepEqual working
			marshalledWant, _ := json.Marshal(&tt.want)
			marshalledGot, _ := json.Marshal(got)

			var data1, data2 any

			_ = json.Unmarshal(marshalledWant, &data1)
			_ = json.Unmarshal(marshalledGot, &data2)

			assert.Equal(t, data1, data2, "The unmarshalled data structures are not equal")
		})
	}
}

func Test_computeBillingHistory(t *testing.T) {
	history := []historyHolder{
		{
			billingHistory: shared.BillingHistory{
				Date: shared.NewDate("2020-01-01"),
				Event: shared.TransactionEvent{
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeReappliedCredit},
				},
			},
			balanceAdjustment: -500,
			creditAdjustment:  -500,
		},
		{
			billingHistory: shared.BillingHistory{
				Date: shared.NewDate("2020-01-01"),
				Event: shared.TransactionEvent{
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceAdjustmentApplied},
				},
			},
			balanceAdjustment: 500,
		},
		{
			billingHistory: shared.BillingHistory{
				Date: shared.NewDate("2021-01-01"),
				Event: shared.InvoiceGenerated{
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceGenerated},
				},
			},
			balanceAdjustment: 32000,
		},
		{
			billingHistory: shared.BillingHistory{
				Date: shared.NewDate("2025-01-01"),
				Event: shared.InvoiceGenerated{
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceGenerated},
				},
			},
			balanceAdjustment: 10000,
		},
		{
			billingHistory: shared.BillingHistory{
				Date: shared.NewDate("2024-01-01"),
				Event: shared.FeeReductionCancelled{
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionCancelled},
				},
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				Date: shared.NewDate("2023-01-01"),
				Event: shared.FeeReductionAwarded{
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionAwarded},
				},
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
			OutstandingBalance: 42000,
			CreditBalance:      -500,
		},
		{
			Date: shared.NewDate("2024-01-01"),
			Event: shared.FeeReductionCancelled{
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionCancelled},
			},
			OutstandingBalance: 32000,
			CreditBalance:      -500,
		},
		{
			Date: shared.NewDate("2023-01-01"),
			Event: shared.FeeReductionAwarded{
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionAwarded},
			},
			OutstandingBalance: 32000,
			CreditBalance:      -500,
		},
		{
			Date: shared.NewDate("2021-01-01"),
			Event: shared.InvoiceGenerated{
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceGenerated},
			},
			OutstandingBalance: 32000,
			CreditBalance:      -500,
		},
		{
			Date: shared.NewDate("2020-01-01"),
			Event: shared.TransactionEvent{
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeReappliedCredit},
			},
			OutstandingBalance: 0,
			CreditBalance:      -500,
		},
		{
			Date: shared.NewDate("2020-01-01"),
			Event: shared.TransactionEvent{
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceAdjustmentApplied},
			},
			OutstandingBalance: 500,
		},
	}

	billingHistory := computeBillingHistory(history)
	assert.Equalf(t, expected, billingHistory, "computeBillingHistory(%v)", history)
}

func Test_invoiceEvents(t *testing.T) {
	invoices := []store.GetGeneratedInvoicesRow{
		{
			InvoiceID: 1,
			Reference: "AD123455/01",
			Feetype:   "AD",
			Amount:    100000,
			CreatedBy: pgtype.Int4{Int32: 3, Valid: true},
			CreatedAt: pgtype.Timestamp{Time: time.Date(2027, time.March, 31, 0, 0, 0, 0, time.UTC), Valid: true},
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
			Notes:              "Awarded 1",
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
			Notes:              "Awarded 2",
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
					Notes:            "Awarded 1",
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
		{
			billingHistory: shared.BillingHistory{
				User: 1,
				Date: shared.Date{Time: now.Add(72 * time.Hour)},
				Event: shared.FeeReductionAwarded{
					ReductionType:    shared.FeeReductionTypeRemission,
					StartDate:        shared.Date{Time: now},
					EndDate:          shared.Date{Time: now.Add(24 * time.Hour)},
					DateReceived:     shared.Date{Time: now.Add(48 * time.Hour)},
					Notes:            "Awarded 2",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionAwarded},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
	}

	assert.Equalf(t, expected, processFeeReductionEvents(reductions), "processFeeReductionEvents(%v)", reductions)
}

func Test_processLedgerAllocations(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		allocations []store.GetLedgerAllocationsForClientRow
		clientID    int32
		want        []historyHolder
	}{
		{
			name:        "No allocations",
			allocations: []store.GetLedgerAllocationsForClientRow{},
			clientID:    1,
			want:        nil,
		},
		{
			name: "Unapply",
			allocations: []store.GetLedgerAllocationsForClientRow{
				{
					LedgerID:         1,
					InvoiceID:        pgtype.Int4{Int32: 2, Valid: true},
					Reference:        pgtype.Text{String: "abc1/23", Valid: true},
					Type:             "CREDIT MEMO",
					Status:           "ALLOCATED",
					LedgerAmount:     5000,
					AllocationAmount: 10000,
					CreatedAt: pgtype.Timestamp{
						Time:  now,
						Valid: true,
					},
					CreatedBy: pgtype.Int4{
						Int32: 3,
						Valid: true,
					},
				},
				{
					LedgerID:         1,
					InvoiceID:        pgtype.Int4{Int32: 2, Valid: true},
					Reference:        pgtype.Text{String: "abc1/23", Valid: true},
					Type:             "CREDIT MEMO",
					Status:           "UNAPPLIED",
					LedgerAmount:     5000,
					AllocationAmount: -5000,
					CreatedAt: pgtype.Timestamp{
						Time:  now,
						Valid: true,
					},
					CreatedBy: pgtype.Int4{
						Int32: 3,
						Valid: true,
					},
				},
			},
			clientID: 99,
			want: []historyHolder{
				{
					billingHistory: shared.BillingHistory{
						User: 3,
						Date: shared.Date{Time: now},
						Event: shared.TransactionEvent{
							ClientId:        99,
							TransactionType: shared.TransactionTypeCreditMemo,
							Amount:          10000,
							Breakdown: []shared.PaymentBreakdown{
								{
									InvoiceReference: shared.InvoiceEvent{
										ID:        2,
										Reference: "abc1/23",
									},
									Amount: 10000,
									Status: "ALLOCATED",
								},
								{
									InvoiceReference: shared.InvoiceEvent{
										ID:        2,
										Reference: "abc1/23",
									},
									Amount: 5000,
									Status: "UNAPPLIED",
								},
							},
							BaseBillingEvent: shared.BaseBillingEvent{
								Type: shared.EventTypeInvoiceAdjustmentApplied,
							},
						},
					},
					balanceAdjustment: -5000,
					creditAdjustment:  5000,
				},
			},
		},
		{
			name: "Reapply",
			allocations: []store.GetLedgerAllocationsForClientRow{
				{
					LedgerID:         1,
					InvoiceID:        pgtype.Int4{Int32: 2, Valid: true},
					Reference:        pgtype.Text{String: "abc1/23", Valid: true},
					Type:             "CREDIT REAPPLY",
					Status:           "REAPPLIED",
					LedgerAmount:     5000,
					AllocationAmount: 5000,
					CreatedAt: pgtype.Timestamp{
						Time:  now,
						Valid: true,
					},
					CreatedBy: pgtype.Int4{
						Int32: 3,
						Valid: true,
					},
				},
			},
			clientID: 99,
			want: []historyHolder{
				{
					billingHistory: shared.BillingHistory{
						User: 3,
						Date: shared.Date{Time: now},
						Event: shared.TransactionEvent{
							ClientId:        99,
							TransactionType: shared.TransactionTypeReapply,
							Amount:          5000,
							Breakdown: []shared.PaymentBreakdown{
								{
									InvoiceReference: shared.InvoiceEvent{
										ID:        2,
										Reference: "abc1/23",
									},
									Amount: 5000,
									Status: "REAPPLIED",
								},
							},
							BaseBillingEvent: shared.BaseBillingEvent{
								Type: shared.EventTypeReappliedCredit,
							},
						},
					},
					balanceAdjustment: -5000,
					creditAdjustment:  -5000,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, processLedgerAllocations(tt.allocations, tt.clientID), "processLedgerAllocations(%v, %v)", tt.allocations, tt.clientID)
		})
	}
}
