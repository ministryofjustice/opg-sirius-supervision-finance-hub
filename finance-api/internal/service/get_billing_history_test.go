package service

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) TestService_GetBillingHistory() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO supervision_finance.finance_client VALUES (1,11,1234,'DEMANDED',NULL);",

		// event 1 - invoice created
		"INSERT INTO supervision_finance.invoice VALUES (1,11,1,'AD','AD000001/24','2024-01-01','2024-01-01',10000,NULL,'2024-01-01',NULL,'2024-01-01','Created manually',NULL,NULL,'2024-01-01 09:31:44',1);",
		// event 2 - direct debit mandate created
		"INSERT INTO supervision_finance.payment_method VALUES (1, 1, 'DIRECT DEBIT', '2024-01-02 00:00:00', 1);",
		// event 3 - invoice adjustment created, event 4 - invoice adjustment rejected
		"INSERT INTO supervision_finance.invoice_adjustment VALUES (1,1,1,'2024-01-03','CREDIT WRITE OFF',10000,'Writing off','REJECTED','2024-01-03 09:32:23',1,'2024-01-04 09:33:24',1)",
		// event 5 - invoice adjustment created, event 6 - invoice adjustment applied
		"INSERT INTO supervision_finance.invoice_adjustment VALUES (5,1,1,'2024-01-05','CREDIT MEMO',8000,'Adding credit','APPROVED','2024-01-05 09:34:38',1,'2024-01-06 09:34:44',1)",
		"INSERT INTO supervision_finance.ledger VALUES (1,'adjustment-ledger','2024-01-06 09:34:44','',8000,'Credit due to approved credit memo','CREDIT MEMO','CONFIRMED',1,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2024-01-06 09:34:44',1);",
		"INSERT INTO supervision_finance.ledger_allocation VALUES (1,1,1,'2024-01-06 09:34:44',8000,'ALLOCATED');",
		// event 7 - fee reduction awarded, event 8 - fee reduction applied and creates unapply, event 9 - fee reduction cancelled
		"INSERT INTO supervision_finance.fee_reduction VALUES (1,1,'REMISSION',NULL,'2024-01-07','2027-03-31','Needs remission',TRUE,'2024-01-07','2024-01-07 09:32:50',1,'2024-01-09 09:33:19',1,'Wrong remission');",
		"INSERT INTO supervision_finance.ledger VALUES (2,'remission-ledger','2024-01-08 09:34:44','',2000,'Credit due to approved remission','CREDIT REMISSION','CONFIRMED',1,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2024-04-08 09:32:50',1);",
		"INSERT INTO supervision_finance.ledger_allocation VALUES (2,2,1,'2024-01-08 09:32:50',10000,'ALLOCATED');",
		"INSERT INTO supervision_finance.ledger_allocation VALUES (3,2,1,'2024-01-08 09:32:50',-8000,'UNAPPLIED');",
		// event 10 - invoice created, event 11 credit reapplied
		"INSERT INTO supervision_finance.invoice VALUES (2,11,1,'AD','AD000002/24','2024-01-10','2024-01-10',10000,NULL,'2024-01-10',NULL,'2024-01-10','Created manually',NULL,NULL,'2024-01-10 09:35:03',1);",
		"INSERT INTO supervision_finance.ledger VALUES (3,'invoice-reapplied','2024-01-11 09:35:03','',8000,'Excess credit applied to invoice','CREDIT REAPPLY','CONFIRMED',1,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2024-01-11 09:35:03',1);",
		"INSERT INTO supervision_finance.ledger_allocation VALUES (4,3,2,'2024-01-11 09:35:03',8000,'REAPPLIED');",
		// event 12 - dd schedule created, event 13 - dd collected with overpayment
		"INSERT INTO supervision_finance.pending_collection VALUES (1, 1, '2024-01-13', 10000, 'COLLECTED', NULL, '2024-01-12 00:00:00', 1);",
		"INSERT INTO supervision_finance.ledger VALUES (4,'dd-ledger','2024-01-13 09:36:05','',10000,'DD payment','DIRECT DEBIT PAYMENT','CONFIRMED',1,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2024-01-13 09:36:05',1);",
		"INSERT INTO supervision_finance.ledger_allocation VALUES (5,4,2,'2024-01-13 09:36:05',2000,'ALLOCATED');",
		"INSERT INTO supervision_finance.ledger_allocation VALUES (6,4,NULL,'2024-01-13 09:36:05',-8000,'UNAPPLIED');",
		// event 14 - refund created, event 15 - refund approved, event 16 - refund processing, event 17 - refund processed
		"INSERT INTO supervision_finance.refund VALUES (1, 1, '2024-01-14', 8000, 'APPROVED', 'refund needed', 1, '2024-01-14', 2, '2024-01-15', '2024-01-16', NULL, '2024-01-17');",
		"INSERT INTO supervision_finance.ledger VALUES (5,'refund-ledger','2024-01-17 09:36:05','',-8000,'refund','REFUND','CONFIRMED',1,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2024-01-17 09:36:05',1);",
		"INSERT INTO supervision_finance.ledger_allocation VALUES (7,5,NULL,'2024-01-17 09:36:05',8000,'REAPPLIED');",

		// legacy events - not to be displayed
		"INSERT INTO supervision_finance.fee_reduction VALUES (22, 1, 'HARDSHIP', NULL, '2019-04-01', '2020-03-31', 'Legacy (no created BankDate) - do not display', FALSE, '2019-05-01');",
		// events that are for a different client - not to be displayed
		"INSERT INTO supervision_finance.finance_client VALUES (2,22,2234,'DEMANDED',NULL);",
		"INSERT INTO supervision_finance.invoice VALUES (3,22,2,'AD','AD000003/24','2024-01-01','2024-01-01',15000,NULL,'2024-01-01',NULL,'2024-01-01','Created manually',NULL,NULL,'2024-01-01 10:00:00',2);",
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
			id:   11, // client id different from finance_client.id to test joins
			want: []shared.BillingHistory{
				{
					User: 1,
					Date: shared.NewDate("2024-01-17"),
					Event: shared.TransactionEvent{
						ClientId:        11,
						TransactionType: shared.TransactionTypeRefund,
						Amount:          -8000,
						Breakdown: []shared.PaymentBreakdown{
							{
								Amount: 8000,
								Status: "REAPPLIED",
							},
						},
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeRefundProcessed,
						},
					},
					OutstandingBalance: 0,
					CreditBalance:      0,
				},
				{
					User: 2,
					Date: shared.NewDate("2024-01-16"),
					Event: shared.RefundEvent{
						ClientId: 11,
						Id:       1,
						Amount:   8000,
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeRefundProcessing,
						},
						Notes: "refund needed",
					},
					OutstandingBalance: 0,
					CreditBalance:      8000,
				},
				{
					User: 2,
					Date: shared.NewDate("2024-01-15"),
					Event: shared.RefundEvent{
						ClientId: 11,
						Id:       1,
						Amount:   8000,
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeRefundApproved,
						},
						Notes: "refund needed",
					},
					OutstandingBalance: 0,
					CreditBalance:      8000,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-01-14"),
					Event: shared.RefundEvent{
						ClientId: 11,
						Id:       1,
						Amount:   8000,
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeRefundCreated,
						},
						Notes: "refund needed",
					},
					OutstandingBalance: 0,
					CreditBalance:      8000,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-01-13 00:00:00"),
					Event: shared.PaymentProcessed{
						TransactionEvent: shared.TransactionEvent{
							ClientId:        11,
							TransactionType: shared.TransactionTypeDirectDebitPayment,
							Amount:          10000,
							Breakdown: []shared.PaymentBreakdown{
								{
									InvoiceReference: shared.InvoiceEvent{
										ID:        2,
										Reference: "AD000002/24",
									},
									Amount: 2000,
									Status: "ALLOCATED",
								},
								{
									Amount: 8000,
									Status: "UNAPPLIED",
								},
							},
							BaseBillingEvent: shared.BaseBillingEvent{
								Type: shared.EventTypePaymentProcessed,
							},
						},
					},
					OutstandingBalance: 0,
					CreditBalance:      8000,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-01-12 00:00:00"),
					Event: shared.DirectDebitEvent{
						Amount:         10000,
						CollectionDate: shared.NewDate("2024-01-13 00:00:00"),
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeDirectDebitCollectionScheduled,
						},
					},
					OutstandingBalance: 2000,
					CreditBalance:      0,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-01-11"),
					Event: shared.TransactionEvent{
						ClientId:        11,
						TransactionType: shared.TransactionTypeReapply,
						Amount:          8000,
						Breakdown: []shared.PaymentBreakdown{
							{
								InvoiceReference: shared.InvoiceEvent{
									ID:        2,
									Reference: "AD000002/24",
								},
								Amount: 8000,
								Status: "REAPPLIED",
							},
						},
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeReappliedCredit,
						},
					},
					OutstandingBalance: 2000,
					CreditBalance:      0,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-01-10"),
					Event: shared.InvoiceGenerated{
						ClientId: 11,
						InvoiceReference: shared.InvoiceEvent{
							ID:        2,
							Reference: "AD000002/24",
						},
						InvoiceType: shared.InvoiceTypeAD,
						Amount:      10000,
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeInvoiceGenerated,
						},
					},
					OutstandingBalance: 10000,
					CreditBalance:      8000,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-01-09"),
					Event: shared.FeeReductionCancelled{
						ReductionType:      shared.FeeReductionTypeRemission,
						CancellationReason: "Wrong remission",
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeFeeReductionCancelled,
						},
					},
					OutstandingBalance: 0,
					CreditBalance:      8000,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-01-08"),
					Event: shared.FeeReductionApplied{
						TransactionEvent: shared.TransactionEvent{
							ClientId:        11,
							TransactionType: shared.TransactionTypeRemission,
							Amount:          10000,
							Breakdown: []shared.PaymentBreakdown{
								{
									InvoiceReference: shared.InvoiceEvent{
										ID:        1,
										Reference: "AD000001/24",
									},
									Amount: 10000,
									Status: "ALLOCATED",
								},
								{
									InvoiceReference: shared.InvoiceEvent{
										ID:        1,
										Reference: "AD000001/24",
									},
									Amount: 8000,
									Status: "UNAPPLIED",
								},
							},
							BaseBillingEvent: shared.BaseBillingEvent{
								Type: shared.EventTypeFeeReductionApplied,
							},
						},
					},
					OutstandingBalance: 0,
					CreditBalance:      8000,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-01-07"),
					Event: shared.FeeReductionAwarded{
						ReductionType: shared.FeeReductionTypeRemission,
						StartDate:     shared.NewDate("2024-01-07"),
						EndDate:       shared.NewDate("2027-03-31"),
						DateReceived:  shared.NewDate("2024-01-07"),
						Notes:         "Needs remission",
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeFeeReductionAwarded,
						},
					},
					OutstandingBalance: 2000,
					CreditBalance:      0,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-01-06"),
					Event: shared.InvoiceAdjustmentApplied{
						TransactionEvent: shared.TransactionEvent{
							ClientId:        11,
							TransactionType: shared.TransactionTypeCreditMemo,
							Amount:          8000,
							Breakdown: []shared.PaymentBreakdown{
								{
									InvoiceReference: shared.InvoiceEvent{
										ID:        1,
										Reference: "AD000001/24",
									},
									Amount: 8000,
									Status: "ALLOCATED",
								},
							},
							BaseBillingEvent: shared.BaseBillingEvent{
								Type: shared.EventTypeInvoiceAdjustmentApplied,
							},
						},
					},
					OutstandingBalance: 2000,
					CreditBalance:      0,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-01-05"),
					Event: shared.InvoiceAdjustmentPending{
						AdjustmentType: shared.AdjustmentTypeCreditMemo,
						ClientId:       11,
						Notes:          "Adding credit",
						PaymentBreakdown: shared.PaymentBreakdown{
							InvoiceReference: shared.InvoiceEvent{
								ID:        1,
								Reference: "AD000001/24",
							},
							Amount: 8000,
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
					Date: shared.NewDate("2024-01-04"),
					Event: shared.InvoiceAdjustmentRejected{
						AdjustmentType: shared.AdjustmentTypeWriteOff,
						ClientId:       11,
						Notes:          "Writing off",
						PaymentBreakdown: shared.PaymentBreakdown{
							InvoiceReference: shared.InvoiceEvent{
								ID:        1,
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
					Date: shared.NewDate("2024-01-03"),
					Event: shared.InvoiceAdjustmentPending{
						AdjustmentType: shared.AdjustmentTypeWriteOff,
						ClientId:       11,
						Notes:          "Writing off",
						PaymentBreakdown: shared.PaymentBreakdown{
							InvoiceReference: shared.InvoiceEvent{
								ID:        1,
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
					Date: shared.NewDate("2024-01-02"),
					Event: shared.PaymentMethodChangedEvent{
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeDirectDebitMandateCreated,
						},
					},
					OutstandingBalance: 10000,
					CreditBalance:      0,
				},
				{
					User: 1,
					Date: shared.NewDate("2024-01-01"),
					Event: shared.InvoiceGenerated{
						ClientId: 11,
						InvoiceReference: shared.InvoiceEvent{
							ID:        1,
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
			id:   99,
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
		// payment received to credit
		{
			billingHistory: shared.BillingHistory{
				Date: shared.NewDate("2020-01-01"),
				Event: shared.TransactionEvent{
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypePaymentProcessed},
				},
			},
			creditAdjustment: 500,
		},
		// invoice created
		{
			billingHistory: shared.BillingHistory{
				Date: shared.NewDate("2021-01-01"),
				Event: shared.InvoiceGenerated{
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceGenerated},
				},
			},
			balanceAdjustment: 1000,
		},
		// reapplied credit
		{
			billingHistory: shared.BillingHistory{
				Date: shared.NewDate("2022-01-01"),
				Event: shared.TransactionEvent{
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeReappliedCredit},
				},
			},
			balanceAdjustment: -500,
			creditAdjustment:  -500,
		},
		// fee reduction applied
		{
			billingHistory: shared.BillingHistory{
				Date: shared.NewDate("2023-01-01"),
				Event: shared.TransactionEvent{
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionApplied},
				},
			},
			balanceAdjustment: -500,
		},
		// non-balance affecting events
		{
			billingHistory: shared.BillingHistory{
				Date: shared.NewDate("2022-06-01"),
				Event: shared.FeeReductionAwarded{
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionAwarded},
				},
			},
		},
		{
			billingHistory: shared.BillingHistory{
				Date: shared.NewDate("2021-06-01"),
				Event: shared.PaymentMethodChangedEvent{
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeDirectDebitMandateCreated},
				},
			},
		},
	}

	expected := []shared.BillingHistory{
		{
			Date: shared.NewDate("2023-01-01"),
			Event: shared.TransactionEvent{
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionApplied},
			},
			OutstandingBalance: 0,
			CreditBalance:      0,
		},
		{
			Date: shared.NewDate("2022-06-01"),
			Event: shared.FeeReductionAwarded{
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeFeeReductionAwarded},
			},
			OutstandingBalance: 500,
			CreditBalance:      0,
		},
		{
			Date: shared.NewDate("2022-01-01"),
			Event: shared.TransactionEvent{
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeReappliedCredit},
			},
			OutstandingBalance: 500,
			CreditBalance:      0,
		},
		{
			Date: shared.NewDate("2021-06-01"),
			Event: shared.PaymentMethodChangedEvent{
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeDirectDebitMandateCreated},
			},
			OutstandingBalance: 1000,
			CreditBalance:      500,
		},
		{
			Date: shared.NewDate("2021-01-01"),
			Event: shared.InvoiceGenerated{
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceGenerated},
			},
			OutstandingBalance: 1000,
			CreditBalance:      500,
		},
		{
			Date: shared.NewDate("2020-01-01"),
			Event: shared.TransactionEvent{
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypePaymentProcessed},
			},
			CreditBalance: 500,
		},
	}

	// randomise order to ensure sorting works
	rand.Shuffle(len(history), func(i, j int) { history[i], history[j] = history[j], history[i] })
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
					LedgerDatetime: pgtype.Timestamp{
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
					LedgerDatetime: pgtype.Timestamp{
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
					LedgerDatetime: pgtype.Timestamp{
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
		{
			name: "Payment",
			allocations: []store.GetLedgerAllocationsForClientRow{
				{
					LedgerID:         1,
					InvoiceID:        pgtype.Int4{Int32: 4, Valid: true},
					Reference:        pgtype.Text{String: "def1/24", Valid: true},
					Type:             "OPG BACS PAYMENT",
					Status:           "ALLOCATED",
					LedgerAmount:     5000,
					AllocationAmount: 1000,
					LedgerDatetime: pgtype.Timestamp{
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
					InvoiceID:        pgtype.Int4{Int32: 5, Valid: true},
					Reference:        pgtype.Text{String: "def2/24", Valid: true},
					Type:             "OPG BACS PAYMENT",
					Status:           "ALLOCATED",
					LedgerAmount:     5000,
					AllocationAmount: 4000,
					LedgerDatetime: pgtype.Timestamp{
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
						Event: shared.TransactionEvent{
							ClientId:        99,
							TransactionType: shared.TransactionTypeOPGBACSPayment,
							Amount:          5000,
							Breakdown: []shared.PaymentBreakdown{
								{
									InvoiceReference: shared.InvoiceEvent{
										ID:        4,
										Reference: "def1/24",
									},
									Amount: 1000,
									Status: "ALLOCATED",
								},
								{
									InvoiceReference: shared.InvoiceEvent{
										ID:        5,
										Reference: "def2/24",
									},
									Amount: 4000,
									Status: "ALLOCATED",
								},
							},
							BaseBillingEvent: shared.BaseBillingEvent{
								Type: shared.EventTypePaymentProcessed,
							},
						},
					},
					balanceAdjustment: -5000,
					creditAdjustment:  0,
				},
			},
		},
		{
			name: "Refund",
			allocations: []store.GetLedgerAllocationsForClientRow{
				{
					LedgerID:         1,
					Type:             "REFUND",
					Status:           "REAPPLIED",
					LedgerAmount:     -5000,
					AllocationAmount: 5000,
					LedgerDatetime: pgtype.Timestamp{
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
						Event: shared.TransactionEvent{
							ClientId:        99,
							TransactionType: shared.TransactionTypeRefund,
							Amount:          -5000,
							Breakdown: []shared.PaymentBreakdown{
								{
									Amount: 5000,
									Status: "REAPPLIED",
								},
							},
							BaseBillingEvent: shared.BaseBillingEvent{
								Type: shared.EventTypeRefundProcessed,
							},
						},
					},
					balanceAdjustment: 0000,
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

func Test_getRefundEventTypeAndDate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		refund        store.GetRefundsForBillingHistoryRow
		wantEventType shared.BillingEventType
		wantEventDate time.Time
	}{
		{
			name: "Pending refund returns refund created and raised date",
			refund: store.GetRefundsForBillingHistoryRow{
				RefundID:    1,
				RaisedDate:  pgtype.Date{Time: now, Valid: true},
				Amount:      23,
				Decision:    "PENDING",
				Notes:       "Pending timeline event",
				CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
				CreatedBy:   2,
				DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(1 * time.Hour), Valid: true}),
				DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
				ProcessedAt: pgtype.Timestamp{},
				CancelledAt: pgtype.Timestamp{},
				FulfilledAt: pgtype.Timestamp{},
				CancelledBy: pgtype.Int4{},
			},
			wantEventType: shared.EventTypeRefundCreated,
			wantEventDate: now,
		},
		{
			name: "Rejected refund returns refund status updated and decision at date",
			refund: store.GetRefundsForBillingHistoryRow{
				RefundID:    1,
				RaisedDate:  pgtype.Date{Time: now, Valid: true},
				Amount:      23,
				Decision:    "REJECTED",
				Notes:       "Rejected timeline event",
				CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
				CreatedBy:   2,
				DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true}),
				DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
				ProcessedAt: pgtype.Timestamp{},
				CancelledAt: pgtype.Timestamp{},
				FulfilledAt: pgtype.Timestamp{},
				CancelledBy: pgtype.Int4{},
			},
			wantEventType: shared.EventTypeRefundStatusUpdated,
			wantEventDate: now.Add(24 * time.Hour),
		},
		{
			name: "Cancelled refund - cancelled at approval stage - returns refund cancelled and cancelled at date",
			refund: store.GetRefundsForBillingHistoryRow{
				RefundID:    1,
				RaisedDate:  pgtype.Date{Time: now, Valid: true},
				Amount:      23,
				Decision:    "APPROVED",
				Notes:       "Approved refund then cancelled timeline event",
				CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
				CreatedBy:   2,
				DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true}),
				DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
				ProcessedAt: pgtype.Timestamp(pgtype.Date{}),
				CancelledAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(72 * time.Hour), Valid: true}),
				FulfilledAt: pgtype.Timestamp(pgtype.Date{}),
				CancelledBy: pgtype.Int4{Int32: 2, Valid: true},
			},
			wantEventType: shared.EventTypeRefundCancelled,
			wantEventDate: now.Add(72 * time.Hour),
		},
		{
			name: "Cancelled refund - cancelled at processing stage - returns refund cancelled and cancelled at date",
			refund: store.GetRefundsForBillingHistoryRow{
				RefundID:    1,
				RaisedDate:  pgtype.Date{Time: now, Valid: true},
				Amount:      23,
				Decision:    "APPROVED",
				Notes:       "Processing then cancelled timeline event",
				CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
				CreatedBy:   2,
				DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true}),
				DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
				ProcessedAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(48 * time.Hour), Valid: true}),
				CancelledAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(72 * time.Hour), Valid: true}),
				FulfilledAt: pgtype.Timestamp(pgtype.Date{}),
				CancelledBy: pgtype.Int4{Int32: 2, Valid: true},
			},
			wantEventType: shared.EventTypeRefundCancelled,
			wantEventDate: now.Add(72 * time.Hour),
		},
		{
			name: "Approved refund with processed at date returns refund processing and processing date",
			refund: store.GetRefundsForBillingHistoryRow{
				RefundID:    1,
				RaisedDate:  pgtype.Date{Time: now, Valid: true},
				Amount:      23,
				Decision:    "APPROVED",
				Notes:       "Approved timeline event (with processing date)",
				CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
				CreatedBy:   2,
				DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true}),
				DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
				ProcessedAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(48 * time.Hour), Valid: true}),
				CancelledAt: pgtype.Timestamp(pgtype.Date{}),
				FulfilledAt: pgtype.Timestamp(pgtype.Date{}),
				CancelledBy: pgtype.Int4{Int32: 2, Valid: true},
			},
			wantEventType: shared.EventTypeRefundProcessing,
			wantEventDate: now.Add(48 * time.Hour),
		},
		{
			name: "Approved refund without processed at date returns refund approved and decision at date",
			refund: store.GetRefundsForBillingHistoryRow{
				RefundID:    1,
				RaisedDate:  pgtype.Date{Time: now, Valid: true},
				Amount:      23,
				Decision:    "APPROVED",
				Notes:       "Approved timeline event (without processing date)",
				CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
				CreatedBy:   2,
				DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true}),
				DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
				ProcessedAt: pgtype.Timestamp(pgtype.Date{}),
				CancelledAt: pgtype.Timestamp(pgtype.Date{}),
				FulfilledAt: pgtype.Timestamp(pgtype.Date{}),
				CancelledBy: pgtype.Int4{Int32: 2, Valid: true},
			},
			wantEventType: shared.EventTypeRefundApproved,
			wantEventDate: now.Add(24 * time.Hour),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualEventType, actualDate := getRefundEventTypeAndDate(tt.refund)
			assert.Equalf(t, tt.wantEventType, actualEventType, "getRefundEventType(%v, %v)", tt.wantEventType, actualEventType)
			assert.Equalf(t, tt.wantEventDate, actualDate, "getRefundEventDate(%v, %v)", tt.wantEventDate, actualDate)
		})
	}
}

func Test_getUserForEventType(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		refund         store.GetRefundsForBillingHistoryRow
		eventType      shared.BillingEventType
		expectedResult int32
	}{
		{
			name: "Pending refund returns created user",
			refund: store.GetRefundsForBillingHistoryRow{
				RefundID:    1,
				RaisedDate:  pgtype.Date{Time: now, Valid: true},
				Amount:      23,
				Decision:    "PENDING",
				Notes:       "Pending timeline event",
				CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
				CreatedBy:   2,
				DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(1 * time.Hour), Valid: true}),
				DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
				ProcessedAt: pgtype.Timestamp{},
				CancelledAt: pgtype.Timestamp{},
				FulfilledAt: pgtype.Timestamp{},
				CancelledBy: pgtype.Int4{},
			},
			eventType:      shared.EventTypeRefundCreated,
			expectedResult: 2,
		},
		{
			name: "Rejected refund returns decision by user",
			refund: store.GetRefundsForBillingHistoryRow{
				RefundID:    1,
				RaisedDate:  pgtype.Date{Time: now, Valid: true},
				Amount:      23,
				Decision:    "REJECTED",
				Notes:       "Rejected timeline event",
				CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
				CreatedBy:   2,
				DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true}),
				DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
				ProcessedAt: pgtype.Timestamp{},
				CancelledAt: pgtype.Timestamp{},
				FulfilledAt: pgtype.Timestamp{},
				CancelledBy: pgtype.Int4{},
			},
			eventType:      shared.EventTypeRefundStatusUpdated,
			expectedResult: 1,
		},
		{
			name: "Cancelled refund - cancelled at approval stage - returns cancelled by user",
			refund: store.GetRefundsForBillingHistoryRow{
				RefundID:    1,
				RaisedDate:  pgtype.Date{Time: now, Valid: true},
				Amount:      23,
				Decision:    "APPROVED",
				Notes:       "Approved refund then cancelled timeline event",
				CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
				CreatedBy:   2,
				DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true}),
				DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
				ProcessedAt: pgtype.Timestamp(pgtype.Date{}),
				CancelledAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(72 * time.Hour), Valid: true}),
				FulfilledAt: pgtype.Timestamp(pgtype.Date{}),
				CancelledBy: pgtype.Int4{Int32: 3, Valid: true},
			},
			eventType:      shared.EventTypeRefundCancelled,
			expectedResult: 3,
		},
		{
			name: "Cancelled refund - cancelled at processing stage - returns cancelled by user",
			refund: store.GetRefundsForBillingHistoryRow{
				RefundID:    1,
				RaisedDate:  pgtype.Date{Time: now, Valid: true},
				Amount:      23,
				Decision:    "APPROVED",
				Notes:       "Processing then cancelled timeline event",
				CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
				CreatedBy:   2,
				DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true}),
				DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
				ProcessedAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(48 * time.Hour), Valid: true}),
				CancelledAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(72 * time.Hour), Valid: true}),
				FulfilledAt: pgtype.Timestamp(pgtype.Date{}),
				CancelledBy: pgtype.Int4{Int32: 3, Valid: true},
			},
			eventType:      shared.EventTypeRefundCancelled,
			expectedResult: 3,
		},
		{
			name: "Approved refund returns decision by user",
			refund: store.GetRefundsForBillingHistoryRow{
				RefundID:    1,
				RaisedDate:  pgtype.Date{Time: now, Valid: true},
				Amount:      23,
				Decision:    "APPROVED",
				Notes:       "Approved timeline event (with processing date)",
				CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
				CreatedBy:   2,
				DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true}),
				DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
				ProcessedAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(48 * time.Hour), Valid: true}),
				CancelledAt: pgtype.Timestamp(pgtype.Date{}),
				FulfilledAt: pgtype.Timestamp(pgtype.Date{}),
				CancelledBy: pgtype.Int4{},
			},
			eventType:      shared.EventTypeRefundApproved,
			expectedResult: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualUser := getUserForEventType(tt.refund, tt.eventType)
			assert.Equalf(t, tt.expectedResult, actualUser, "getUserForEventType(%v, %v)", tt.expectedResult, actualUser)
		})
	}
}

func Test_makeRefundEvent(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		refund         store.GetRefundsForBillingHistoryRow
		user           int32
		eventType      shared.BillingEventType
		date           time.Time
		clientID       int32
		history        []historyHolder
		expectedResult []historyHolder
	}{
		{
			name: "Add event to empty history holder",
			refund: store.GetRefundsForBillingHistoryRow{
				RefundID:    1,
				RaisedDate:  pgtype.Date{Time: now, Valid: true},
				Amount:      23,
				Decision:    "PENDING",
				Notes:       "Pending timeline event",
				CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
				CreatedBy:   2,
				DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(1 * time.Hour), Valid: true}),
				DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
				ProcessedAt: pgtype.Timestamp{},
				CancelledAt: pgtype.Timestamp{},
				FulfilledAt: pgtype.Timestamp{},
				CancelledBy: pgtype.Int4{},
			},
			user:      11,
			eventType: shared.EventTypeRefundCreated,
			date:      now,
			clientID:  45,
			history:   []historyHolder{},
			expectedResult: []historyHolder{
				{
					billingHistory: shared.BillingHistory{
						User: 11,
						Date: shared.Date{Time: now},
						Event: shared.RefundEvent{
							Id:               1,
							ClientId:         45,
							Amount:           23,
							Notes:            "Pending timeline event",
							BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundCreated},
						},
						OutstandingBalance: 0,
					},
					balanceAdjustment: 0,
				},
			},
		},
		{
			name: "Add event to history holder with existing event",
			refund: store.GetRefundsForBillingHistoryRow{
				RefundID:    1,
				RaisedDate:  pgtype.Date{Time: now.Add(1 * time.Hour), Valid: true},
				Amount:      55,
				Decision:    "APPROVED",
				Notes:       "Newer timeline event",
				CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now.Add(1 * time.Hour), Valid: true}),
				CreatedBy:   2,
				DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(12 * time.Hour), Valid: true}),
				DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
				ProcessedAt: pgtype.Timestamp{},
				CancelledAt: pgtype.Timestamp{},
				FulfilledAt: pgtype.Timestamp{},
				CancelledBy: pgtype.Int4{},
			},
			user:      22,
			eventType: shared.EventTypeRefundProcessing,
			date:      now,
			clientID:  66,
			history: []historyHolder{
				{
					billingHistory: shared.BillingHistory{
						User: 11,
						Date: shared.Date{Time: now},
						Event: shared.RefundEvent{
							Id:               1,
							ClientId:         45,
							Amount:           23,
							Notes:            "Existing timeline event",
							BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundStatusUpdated},
						},
						OutstandingBalance: 0,
					},
					balanceAdjustment: 0,
				},
			},
			expectedResult: []historyHolder{
				{
					billingHistory: shared.BillingHistory{
						User: 11,
						Date: shared.Date{Time: now},
						Event: shared.RefundEvent{
							Id:               1,
							ClientId:         45,
							Amount:           23,
							Notes:            "Existing timeline event",
							BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundStatusUpdated},
						},
						OutstandingBalance: 0,
					},
					balanceAdjustment: 0,
				},
				{
					billingHistory: shared.BillingHistory{
						User: 22,
						Date: shared.Date{Time: now},
						Event: shared.RefundEvent{
							Id:               1,
							ClientId:         66,
							Amount:           55,
							Notes:            "Newer timeline event",
							BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundProcessing},
						},
						OutstandingBalance: 0,
					},
					balanceAdjustment: 0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualEvent := makeRefundEvent(tt.refund, tt.user, tt.eventType, tt.date, tt.clientID, tt.history)
			assert.Equalf(t, tt.expectedResult, actualEvent, "makeRefundEvent(%v, %v)", tt.expectedResult, actualEvent)
		})
	}
}

func Test_processRefundEvents(t *testing.T) {
	now := time.Now()
	refunds := []store.GetRefundsForBillingHistoryRow{
		{
			RefundID:    8,
			RaisedDate:  pgtype.Date{Time: now, Valid: true},
			Amount:      23,
			Decision:    "PENDING",
			Notes:       "Pending timeline event",
			CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
			CreatedBy:   2,
			DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(1 * time.Hour), Valid: true}),
			DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
			ProcessedAt: pgtype.Timestamp{},
			CancelledAt: pgtype.Timestamp{},
			FulfilledAt: pgtype.Timestamp{},
			CancelledBy: pgtype.Int4{},
		},
		{
			RefundID:    7,
			RaisedDate:  pgtype.Date{Time: now, Valid: true},
			Amount:      33,
			Decision:    "REJECTED",
			Notes:       "Rejected timeline event",
			CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
			CreatedBy:   2,
			DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true}),
			DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
			ProcessedAt: pgtype.Timestamp{},
			CancelledAt: pgtype.Timestamp{},
			FulfilledAt: pgtype.Timestamp{},
			CancelledBy: pgtype.Int4{},
		},
		{
			RefundID:    6,
			RaisedDate:  pgtype.Date{Time: now, Valid: true},
			Amount:      44,
			Decision:    "APPROVED",
			Notes:       "Fulfilled timeline event", // fulfilled timeline events are created via ledger allocation events, so this will not be included
			CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
			CreatedBy:   2,
			DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true}),
			DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
			ProcessedAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(48 * time.Hour), Valid: true}),
			CancelledAt: pgtype.Timestamp{},
			FulfilledAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(72 * time.Hour), Valid: true}),
			CancelledBy: pgtype.Int4{},
		},
		{
			RefundID:    5,
			RaisedDate:  pgtype.Date{Time: now, Valid: true},
			Amount:      54,
			Decision:    "APPROVED",
			Notes:       "Cancelled timeline event (after being approved)",
			CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
			CreatedBy:   2,
			DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true}),
			DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
			ProcessedAt: pgtype.Timestamp(pgtype.Date{}),
			CancelledAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(72 * time.Hour), Valid: true}),
			FulfilledAt: pgtype.Timestamp(pgtype.Date{}),
			CancelledBy: pgtype.Int4{Int32: 3, Valid: true},
		},
		{
			RefundID:    4,
			RaisedDate:  pgtype.Date{Time: now, Valid: true},
			Amount:      55,
			Decision:    "APPROVED",
			Notes:       "Cancelled timeline event (after being in processing)",
			CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
			CreatedBy:   2,
			DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true}),
			DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
			ProcessedAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(48 * time.Hour), Valid: true}),
			CancelledAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(72 * time.Hour), Valid: true}),
			FulfilledAt: pgtype.Timestamp(pgtype.Date{}),
			CancelledBy: pgtype.Int4{Int32: 3, Valid: true},
		},
		{
			RefundID:    3,
			RaisedDate:  pgtype.Date{Time: now, Valid: true},
			Amount:      66,
			Decision:    "APPROVED",
			Notes:       "Approved timeline event (with processing date)",
			CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
			CreatedBy:   2,
			DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true}),
			DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
			ProcessedAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(48 * time.Hour), Valid: true}),
			CancelledAt: pgtype.Timestamp(pgtype.Date{}),
			FulfilledAt: pgtype.Timestamp(pgtype.Date{}),
			CancelledBy: pgtype.Int4{},
		},
		{
			RefundID:    2,
			RaisedDate:  pgtype.Date{Time: now, Valid: true},
			Amount:      67,
			Decision:    "APPROVED",
			Notes:       "Approved timeline event (without processing date)",
			CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
			CreatedBy:   2,
			DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true}),
			DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
			ProcessedAt: pgtype.Timestamp(pgtype.Date{}),
			CancelledAt: pgtype.Timestamp(pgtype.Date{}),
			FulfilledAt: pgtype.Timestamp(pgtype.Date{}),
			CancelledBy: pgtype.Int4{},
		},
	}

	expected := []historyHolder{
		{
			billingHistory: shared.BillingHistory{
				User: 2,
				Date: shared.Date{Time: now},
				Event: shared.RefundEvent{
					Id:               8,
					ClientId:         33,
					Amount:           23,
					Notes:            "Pending timeline event",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundCreated},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 1,
				Date: shared.Date{Time: now.Add(24 * time.Hour)},
				Event: shared.RefundEvent{
					Id:               7,
					ClientId:         33,
					Amount:           33,
					Notes:            "Rejected timeline event",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundStatusUpdated},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 2,
				Date: shared.Date{Time: now},
				Event: shared.RefundEvent{
					Id:               7,
					ClientId:         33,
					Amount:           33,
					Notes:            "Rejected timeline event",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundCreated},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 1,
				Date: shared.Date{Time: now.Add(48 * time.Hour)},
				Event: shared.RefundEvent{
					Id:               6,
					ClientId:         33,
					Amount:           44,
					Notes:            "Fulfilled timeline event",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundProcessing},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 1,
				Date: shared.Date{Time: now.Add(24 * time.Hour)},
				Event: shared.RefundEvent{
					Id:               6,
					ClientId:         33,
					Amount:           44,
					Notes:            "Fulfilled timeline event",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundApproved},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 2,
				Date: shared.Date{Time: now},
				Event: shared.RefundEvent{
					Id:               6,
					ClientId:         33,
					Amount:           44,
					Notes:            "Fulfilled timeline event",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundCreated},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 3,
				Date: shared.Date{Time: now.Add(72 * time.Hour)},
				Event: shared.RefundEvent{
					Id:               5,
					ClientId:         33,
					Amount:           54,
					Notes:            "Cancelled timeline event (after being approved)",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundCancelled},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 1,
				Date: shared.Date{Time: now.Add(24 * time.Hour)},
				Event: shared.RefundEvent{
					Id:               5,
					ClientId:         33,
					Amount:           54,
					Notes:            "Cancelled timeline event (after being approved)",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundApproved},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 2,
				Date: shared.Date{Time: now},
				Event: shared.RefundEvent{
					Id:               5,
					ClientId:         33,
					Amount:           54,
					Notes:            "Cancelled timeline event (after being approved)",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundCreated},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 3,
				Date: shared.Date{Time: now.Add(72 * time.Hour)},
				Event: shared.RefundEvent{
					Id:               4,
					ClientId:         33,
					Amount:           55,
					Notes:            "Cancelled timeline event (after being in processing)",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundCancelled},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 1,
				Date: shared.Date{Time: now.Add(48 * time.Hour)},
				Event: shared.RefundEvent{
					Id:               4,
					ClientId:         33,
					Amount:           55,
					Notes:            "Cancelled timeline event (after being in processing)",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundProcessing},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 1,
				Date: shared.Date{Time: now.Add(24 * time.Hour)},
				Event: shared.RefundEvent{
					Id:               4,
					ClientId:         33,
					Amount:           55,
					Notes:            "Cancelled timeline event (after being in processing)",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundApproved},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 2,
				Date: shared.Date{Time: now},
				Event: shared.RefundEvent{
					Id:               4,
					ClientId:         33,
					Amount:           55,
					Notes:            "Cancelled timeline event (after being in processing)",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundCreated},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 1,
				Date: shared.Date{Time: now.Add(48 * time.Hour)},
				Event: shared.RefundEvent{
					Id:               3,
					ClientId:         33,
					Amount:           66,
					Notes:            "Approved timeline event (with processing date)",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundProcessing},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 1,
				Date: shared.Date{Time: now.Add(24 * time.Hour)},
				Event: shared.RefundEvent{
					Id:               3,
					ClientId:         33,
					Amount:           66,
					Notes:            "Approved timeline event (with processing date)",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundApproved},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 2,
				Date: shared.Date{Time: now},
				Event: shared.RefundEvent{
					Id:               3,
					ClientId:         33,
					Amount:           66,
					Notes:            "Approved timeline event (with processing date)",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundCreated},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 1,
				Date: shared.Date{Time: now.Add(24 * time.Hour)},
				Event: shared.RefundEvent{
					Id:               2,
					ClientId:         33,
					Amount:           67,
					Notes:            "Approved timeline event (without processing date)",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundApproved},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 2,
				Date: shared.Date{Time: now},
				Event: shared.RefundEvent{
					Id:               2,
					ClientId:         33,
					Amount:           67,
					Notes:            "Approved timeline event (without processing date)",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundCreated},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
	}

	assert.Equalf(t, expected, processRefundEvents(refunds, 33), "processRefundEvents(%v)", refunds)
}

func Test_processPaymentMethodsEvents(t *testing.T) {
	now := time.Now()
	paymentMethods := []store.GetPaymentMethodsForBillingHistoryRow{
		{
			Type:      "DEMANDED",
			CreatedAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(2 * time.Hour), Valid: true}),
			CreatedBy: 1,
		},
		{
			Type:      "DIRECT DEBIT",
			CreatedAt: pgtype.Timestamp{Time: now.Add(72 * time.Hour), Valid: true},
			CreatedBy: 3,
		},
		{
			Type:      "DEMANDED",
			CreatedAt: pgtype.Timestamp{Time: now.Add(180 * time.Hour), Valid: true},
			CreatedBy: 5,
		},
	}

	expected := []historyHolder{
		{
			billingHistory: shared.BillingHistory{
				User: 1,
				Date: shared.Date{Time: now.Add(2 * time.Hour)},
				Event: shared.PaymentMethodChangedEvent{
					BaseBillingEvent: shared.BaseBillingEvent{
						Type: shared.EventTypeDirectDebitMandateCancelled,
					},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 3,
				Date: shared.Date{Time: now.Add(72 * time.Hour)},
				Event: shared.PaymentMethodChangedEvent{
					BaseBillingEvent: shared.BaseBillingEvent{
						Type: shared.EventTypeDirectDebitMandateCreated,
					},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 5,
				Date: shared.Date{Time: now.Add(180 * time.Hour)},
				Event: shared.PaymentMethodChangedEvent{
					BaseBillingEvent: shared.BaseBillingEvent{
						Type: shared.EventTypeDirectDebitMandateCancelled,
					},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
	}

	events := processPaymentMethodEvents(paymentMethods)
	assert.Equalf(t, expected, events, "processPaymentMethodsEvents(%v)", paymentMethods)
}
