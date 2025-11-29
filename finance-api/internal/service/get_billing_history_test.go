package service

import (
	"encoding/json"
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
		"INSERT INTO finance_client VALUES (3,3, 12345,'DEMANDED',NULL);",
		"INSERT INTO invoice VALUES (9,1,1,'AD','AD000001/24','2024-10-07','2024-10-07',10000,NULL,'2024-10-07',NULL,'2024-10-07','Created manually',NULL,NULL,'2024-10-07 09:31:44',1);",
		"INSERT INTO invoice VALUES (10,1,1,'AD','AD000002/24','2024-10-07','2024-10-07',10000,NULL,'2024-10-07',NULL,'2024-10-07','Created manually',NULL,NULL,'2024-10-07 09:35:03',1);",
		"INSERT INTO invoice_adjustment VALUES (4,1,9,'2024-10-07','CREDIT WRITE OFF',10000,'Writing off','REJECTED','2024-10-07 09:32:23',1,'2024-10-07 09:33:24',1)",
		"INSERT INTO invoice_adjustment VALUES (5,1,9,'2024-10-07','CREDIT MEMO',10000,'Adding credit','APPROVED','2024-10-07 09:34:38',1,'2024-10-07 09:34:44',1)",
		"INSERT INTO fee_reduction VALUES (1, 1, 'HARDSHIP', NULL, '2019-04-01', '2020-03-31', 'Legacy (no created BankDate) - do not display', FALSE, '2019-05-01');",
		"INSERT INTO fee_reduction VALUES (5,1,'REMISSION',NULL,'2024-04-01','2027-03-31','Needs remission',TRUE,'2024-10-07','2024-10-07 09:32:50',1,'2024-10-07 09:33:19',1,'Wrong remission');",
		"INSERT INTO ledger VALUES (5,'09799ea2-5f8f-4ecb-8200-f021ab96def1','2024-10-07 09:32:50','',5000,'Credit due to approved remission','CREDIT REMISSION','CONFIRMED',1,NULL,5,NULL,NULL,NULL,NULL,NULL,NULL,NULL,1);",
		"INSERT INTO ledger VALUES (6,'6e469827-fff7-4c22-a2e2-8b7d3580350c','2024-10-07 09:34:44','',5000,'Credit due to approved credit memo','CREDIT MEMO','CONFIRMED',1,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,1);",
		"INSERT INTO ledger VALUES (7,'babda0f7-2f07-4b85-a991-7d45be9474e2','2024-10-07 09:35:03','',5000,'Excess credit applied to invoice','CREDIT REAPPLY','CONFIRMED',1,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,1);",
		"INSERT INTO ledger VALUES (8,'13b3851c-2e7d-43d0-86ad-86ffca586f57','2024-10-07 09:36:05','',1000,'Moto payment','MOTO CARD PAYMENT','CONFIRMED',1,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2024-10-07 09:36:05',1);",

		"INSERT INTO ledger_allocation VALUES (5,5,9,'2024-10-07 09:32:50',5000,'ALLOCATED');",
		"INSERT INTO ledger_allocation VALUES (6,6,9,'2024-10-07 09:34:44',10000,'ALLOCATED');",
		"INSERT INTO ledger_allocation VALUES (7,6,9,'2024-10-07 09:34:44',-5000,'UNAPPLIED',NULL,'Unapplied funds as a result of applying credit memo');",
		"INSERT INTO ledger_allocation VALUES (8,7,10,'2024-10-07 09:35:03',5000,'REAPPLIED');",
		"INSERT INTO ledger_allocation VALUES (9,8,10,'2024-10-07 09:36:05',1000,'ALLOCATED');",

		"INSERT INTO supervision_finance.refund values (16, 3, '2024-01-01', 234, 'REJECTED', 'rejected refund', 1, '2024-07-01', 2, '2024-07-02');",
		"INSERT INTO supervision_finance.refund values (15, 3, '2024-01-01', 234, 'APPROVED', 'processing then cancelled refund', 1, '2024-06-01', 2, '2024-06-02', '2024-06-03', '2024-06-04', null, 3);",
		"INSERT INTO supervision_finance.refund values (14, 3, '2024-01-01', 234, 'APPROVED', 'approved then cancelled refund', 1, '2024-05-01', 2, '2024-05-02', null, '2024-05-03', null, 3);",
		"INSERT INTO supervision_finance.refund values (13, 3, '2024-01-01', 234, 'APPROVED', 'fulfilled refund', 1, '2024-04-01', 2, '2024-04-02', '2024-04-03', null, '2024-04-04');",
		"INSERT INTO supervision_finance.refund values (12, 3, '2024-01-01', 234, 'APPROVED', 'processing refund', 1, '2024-03-01', 2, '2024-03-02', '2024-03-03');",
		"INSERT INTO supervision_finance.refund values (10, 3, '2024-01-01', 234, 'PENDING', 'pending refund', 1, '2024-01-01', null);",
		"INSERT INTO supervision_finance.refund values (11, 3, '2024-01-01', 234, 'APPROVED', 'approved refund', 1, '2024-02-01', 2, '2024-02-02');",

		"INSERT INTO supervision_finance.pending_collection values (10, 1, '2024-01-01', 1233, 'COLLECTED', null, '2023-12-31 00:00:00', 1);",
		"INSERT INTO supervision_finance.pending_collection values (11, 1, '2025-01-01', 333, 'PENDING', null, '2024-12-30 00:00:00', 1);",
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
					Date: shared.NewDate("2024-12-30 00:00:00"),
					Event: shared.DirectDebitEvent{
						Amount:         333,
						CollectionDate: shared.NewDate("2025-01-01 00:00:00"),
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeDirectDebitCollectionScheduled,
						},
					},
					OutstandingBalance: 4000,
					CreditBalance:      0,
				},
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
						StartDate:     shared.NewDate("2024-04-01"),
						EndDate:       shared.NewDate("2027-03-31"),
						DateReceived:  shared.NewDate("2024-10-07"),
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
				{
					User: 1,
					Date: shared.NewDate("2024-01-01 00:00:00"),
					Event: shared.DirectDebitEvent{
						Amount:         1233,
						CollectionDate: shared.NewDate("2024-01-01 00:00:00"),
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeDirectDebitCollected,
						},
					},
					OutstandingBalance: 0,
					CreditBalance:      0,
				},
				{
					User: 1,
					Date: shared.NewDate("2023-12-31 00:00:00"),
					Event: shared.DirectDebitEvent{
						Amount:         1233,
						CollectionDate: shared.NewDate("2024-01-01 00:00:00"),
						BaseBillingEvent: shared.BaseBillingEvent{
							Type: shared.EventTypeDirectDebitCollectionScheduled,
						},
					},
					OutstandingBalance: 0,
					CreditBalance:      0,
				},
			},
		},
		//{
		//	name: "returns correct refund events",
		//	id:   3,
		//	want: []shared.BillingHistory{
		//		{
		//			User: 2,
		//			Date: shared.NewDate("2024-07-02"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       16,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundStatusUpdated,
		//				},
		//				Notes: "rejected refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//		{
		//			User: 1,
		//			Date: shared.NewDate("2024-07-01"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       16,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundCreated,
		//				},
		//				Notes: "rejected refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//		{
		//			User: 3,
		//			Date: shared.NewDate("2024-06-04"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       15,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundCancelled,
		//				},
		//				Notes: "processing then cancelled refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//		{
		//			User: 2,
		//			Date: shared.NewDate("2024-06-03"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       15,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundProcessing,
		//				},
		//				Notes: "processing then cancelled refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//		{
		//			User: 2,
		//			Date: shared.NewDate("2024-06-02"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       15,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundApproved,
		//				},
		//				Notes: "processing then cancelled refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//		{
		//			User: 1,
		//			Date: shared.NewDate("2024-06-01"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       15,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundCreated,
		//				},
		//				Notes: "processing then cancelled refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//		{
		//			User: 3,
		//			Date: shared.NewDate("2024-05-03"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       14,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundCancelled,
		//				},
		//				Notes: "approved then cancelled refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//		{
		//			User: 2,
		//			Date: shared.NewDate("2024-05-02"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       14,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundApproved,
		//				},
		//				Notes: "approved then cancelled refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//		{
		//			User: 1,
		//			Date: shared.NewDate("2024-05-01"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       14,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundCreated,
		//				},
		//				Notes: "approved then cancelled refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//		{
		//			User: 2,
		//			Date: shared.NewDate("2024-04-04"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       13,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundFulfilled,
		//				},
		//				Notes: "fulfilled refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//		{
		//			User: 2,
		//			Date: shared.NewDate("2024-04-03"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       13,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundProcessing,
		//				},
		//				Notes: "fulfilled refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//		{
		//			User: 2,
		//			Date: shared.NewDate("2024-04-02"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       13,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundApproved,
		//				},
		//				Notes: "fulfilled refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//		{
		//			User: 1,
		//			Date: shared.NewDate("2024-04-01"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       13,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundCreated,
		//				},
		//				Notes: "fulfilled refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//		{
		//			User: 2,
		//			Date: shared.NewDate("2024-03-03"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       12,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundProcessing,
		//				},
		//				Notes: "processing refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//		{
		//			User: 2,
		//			Date: shared.NewDate("2024-03-02"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       12,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundApproved,
		//				},
		//				Notes: "processing refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//		{
		//			User: 1,
		//			Date: shared.NewDate("2024-03-01"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       12,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundCreated,
		//				},
		//				Notes: "processing refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//		{
		//			User: 2,
		//			Date: shared.NewDate("2024-02-02"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       11,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundApproved,
		//				},
		//				Notes: "approved refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//		{
		//			User: 1,
		//			Date: shared.NewDate("2024-02-01"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       11,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundCreated,
		//				},
		//				Notes: "approved refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//		{
		//			User: 1,
		//			Date: shared.NewDate("2024-01-01"),
		//			Event: shared.RefundEvent{
		//				ClientId: 3,
		//				Id:       10,
		//				Amount:   234,
		//				BaseBillingEvent: shared.BaseBillingEvent{
		//					Type: shared.EventTypeRefundCreated,
		//				},
		//				Notes: "pending refund",
		//			},
		//			OutstandingBalance: 0,
		//			CreditBalance:      0,
		//		},
		//	},
		//},
		//{
		//	name: "returns an empty array when no match is found",
		//	id:   2,
		//	want: []shared.BillingHistory{},
		//},
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
					Status:           "CONFIRMED",
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
					Status:           "CONFIRMED",
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
									Status: "CONFIRMED",
								},
								{
									InvoiceReference: shared.InvoiceEvent{
										ID:        5,
										Reference: "def2/24",
									},
									Amount: 4000,
									Status: "CONFIRMED",
								},
							},
							BaseBillingEvent: shared.BaseBillingEvent{
								Type: shared.EventTypePaymentProcessed,
							},
						},
					},
					balanceAdjustment: 0,
					creditAdjustment:  0,
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
			name: "Fulfilled refund returns refund fulfilled and fulfilled at date",
			refund: store.GetRefundsForBillingHistoryRow{
				RefundID:    1,
				RaisedDate:  pgtype.Date{Time: now, Valid: true},
				Amount:      23,
				Decision:    "APPROVED",
				Notes:       "Fulfilled timeline event",
				CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
				CreatedBy:   2,
				DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true}),
				DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
				ProcessedAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(48 * time.Hour), Valid: true}),
				CancelledAt: pgtype.Timestamp{},
				FulfilledAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(72 * time.Hour), Valid: true}),
				CancelledBy: pgtype.Int4{},
			},
			wantEventType: shared.EventTypeRefundFulfilled,
			wantEventDate: now.Add(72 * time.Hour),
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
			name: "Fulfilled refund returns decision by user",
			refund: store.GetRefundsForBillingHistoryRow{
				RefundID:    1,
				RaisedDate:  pgtype.Date{Time: now, Valid: true},
				Amount:      23,
				Decision:    "APPROVED",
				Notes:       "Fulfilled timeline event",
				CreatedAt:   pgtype.Timestamp(pgtype.Date{Time: now, Valid: true}),
				CreatedBy:   2,
				DecisionAt:  pgtype.Timestamp(pgtype.Date{Time: now.Add(24 * time.Hour), Valid: true}),
				DecisionBy:  pgtype.Int4{Int32: 1, Valid: true},
				ProcessedAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(48 * time.Hour), Valid: true}),
				CancelledAt: pgtype.Timestamp{},
				FulfilledAt: pgtype.Timestamp(pgtype.Date{Time: now.Add(72 * time.Hour), Valid: true}),
				CancelledBy: pgtype.Int4{},
			},
			eventType:      shared.EventTypeRefundFulfilled,
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

func Test_processRefundEventsCreatesCorrectBillingHistoryEvents(t *testing.T) {
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
			Notes:       "Fulfilled timeline event",
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
				Date: shared.Date{Time: now.Add(72 * time.Hour)},
				Event: shared.RefundEvent{
					Id:               6,
					ClientId:         33,
					Amount:           44,
					Notes:            "Fulfilled timeline event",
					BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeRefundFulfilled},
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

func Test_calculateTotalAmountForPaymentEvents(t *testing.T) {

	tests := []struct {
		name  string
		event shared.TransactionEvent
		want  int
	}{
		{
			name: "No payment breakdown",
			event: shared.TransactionEvent{
				ClientId:        98,
				TransactionType: shared.TransactionTypeOPGBACSPayment,
				Amount:          0,
				Breakdown: []shared.PaymentBreakdown{
					{
						InvoiceReference: shared.InvoiceEvent{},
						Amount:           0,
						Status:           "",
					},
				},
				BaseBillingEvent: shared.BaseBillingEvent{},
			},
			want: 0,
		},
		{
			name: "1 item in breakdown",
			event: shared.TransactionEvent{
				ClientId:        98,
				TransactionType: shared.TransactionTypeOPGBACSPayment,
				Amount:          2222,
				Breakdown: []shared.PaymentBreakdown{
					{
						InvoiceReference: shared.InvoiceEvent{
							ID:        4,
							Reference: "def1/24",
						},
						Amount: 2222,
						Status: "CONFIRMED",
					},
				},
				BaseBillingEvent: shared.BaseBillingEvent{
					Type: shared.EventTypePaymentProcessed,
				},
			},
			want: 2222,
		},
		{
			name: "2 items in breakdown",
			event: shared.TransactionEvent{
				ClientId:        98,
				TransactionType: shared.TransactionTypeOPGBACSPayment,
				Amount:          5000,
				Breakdown: []shared.PaymentBreakdown{
					{
						InvoiceReference: shared.InvoiceEvent{
							ID:        4,
							Reference: "def1/24",
						},
						Amount: 1000,
						Status: "CONFIRMED",
					},
					{
						InvoiceReference: shared.InvoiceEvent{
							ID:        5,
							Reference: "def2/24",
						},
						Amount: 4000,
						Status: "CONFIRMED",
					},
				},
				BaseBillingEvent: shared.BaseBillingEvent{
					Type: shared.EventTypePaymentProcessed,
				},
			},
			want: 5000,
		},
		{
			name: "3 items in breakdown",
			event: shared.TransactionEvent{
				ClientId:        98,
				TransactionType: shared.TransactionTypeOPGBACSPayment,
				Amount:          5432,
				Breakdown: []shared.PaymentBreakdown{
					{
						InvoiceReference: shared.InvoiceEvent{
							ID:        4,
							Reference: "def1/24",
						},
						Amount: 1000,
						Status: "CONFIRMED",
					},
					{
						InvoiceReference: shared.InvoiceEvent{
							ID:        5,
							Reference: "def2/24",
						},
						Amount: 2430,
						Status: "CONFIRMED",
					},
					{
						InvoiceReference: shared.InvoiceEvent{
							ID:        6,
							Reference: "def2/24",
						},
						Amount: 2002,
						Status: "CONFIRMED",
					},
				},
				BaseBillingEvent: shared.BaseBillingEvent{
					Type: shared.EventTypePaymentProcessed,
				},
			},
			want: 5432,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, calculateTotalAmountForPaymentEvents(tt.event), "calculateTotalAmountForPaymentEvents(%v)", tt.event)
		})
	}
}

func Test_makeDirectDebitEvent(t *testing.T) {
	tomorrow := time.Now().AddDate(0, 0, 1)
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	twoDaysAgo := now.AddDate(0, 0, -2)

	tests := []struct {
		name           string
		eventType      shared.BillingEventType
		amount         int32
		user           int32
		createdDate    time.Time
		collectionDate time.Time
		clientID       int32
		history        []historyHolder
		expectedResult []historyHolder
	}{
		{
			name:           "Add event to empty history holder",
			eventType:      shared.EventTypeDirectDebitCollectionScheduled,
			amount:         23,
			user:           11,
			createdDate:    yesterday,
			collectionDate: tomorrow,
			clientID:       45,
			history:        []historyHolder{},
			expectedResult: []historyHolder{
				{
					billingHistory: shared.BillingHistory{
						User: 11,
						Date: shared.Date{Time: yesterday},
						Event: shared.DirectDebitEvent{
							Amount:         23,
							CollectionDate: shared.Date{Time: tomorrow},
							Status:         "",
							BaseBillingEvent: shared.BaseBillingEvent{
								Type: shared.EventTypeDirectDebitCollectionScheduled,
							},
						},
						OutstandingBalance: 0,
					},
					balanceAdjustment: 0,
				},
			},
		},
		{
			name:           "Add collected event to empty history holder",
			eventType:      shared.EventTypeDirectDebitCollected,
			amount:         4444,
			user:           11,
			createdDate:    twoDaysAgo,
			collectionDate: yesterday,
			clientID:       45,
			history:        []historyHolder{},
			expectedResult: []historyHolder{
				{
					billingHistory: shared.BillingHistory{
						User: 11,
						Date: shared.Date{Time: yesterday},
						Event: shared.DirectDebitEvent{
							Amount:         4444,
							CollectionDate: shared.Date{Time: yesterday},
							Status:         "",
							BaseBillingEvent: shared.BaseBillingEvent{
								Type: shared.EventTypeDirectDebitCollected,
							},
						},
						OutstandingBalance: 0,
					},
					balanceAdjustment: 0,
				},
			},
		},
		{
			name:           "Add event to not empty history holder",
			amount:         111,
			user:           11,
			eventType:      shared.EventTypeDirectDebitCollectionFailed,
			createdDate:    now.AddDate(0, 0, -6),
			collectionDate: yesterday,
			clientID:       45,
			history: []historyHolder{
				{
					billingHistory: shared.BillingHistory{
						User: 11,
						Date: shared.Date{Time: now},
						Event: shared.DirectDebitEvent{
							Amount:         23,
							CollectionDate: shared.Date{Time: now.AddDate(0, 0, 2)},
							BaseBillingEvent: shared.BaseBillingEvent{
								Type: shared.EventTypeDirectDebitCollectionScheduled,
							},
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
						Event: shared.DirectDebitEvent{
							Amount:         23,
							CollectionDate: shared.Date{Time: now.AddDate(0, 0, 2)},
							BaseBillingEvent: shared.BaseBillingEvent{
								Type: shared.EventTypeDirectDebitCollectionScheduled,
							},
						},
						OutstandingBalance: 0,
					},
					balanceAdjustment: 0,
				},
				{
					billingHistory: shared.BillingHistory{
						User: 11,
						Date: shared.Date{Time: yesterday},
						Event: shared.DirectDebitEvent{
							Amount:         111,
							CollectionDate: shared.Date{Time: yesterday},
							BaseBillingEvent: shared.BaseBillingEvent{
								Type: shared.EventTypeDirectDebitCollectionFailed,
							},
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
			actualEvent := makeDirectDebitEvent(tt.eventType, tt.amount, tt.user, tt.createdDate, tt.collectionDate, tt.history)
			assert.Equalf(t, tt.expectedResult, actualEvent, "makeDirectDebitEvent(%v, %v)", tt.expectedResult, actualEvent)
		})
	}
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
				Event: shared.BaseBillingEvent{
					Type: shared.EventTypeDirectDebitMandateCancelled,
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 3,
				Date: shared.Date{Time: now.Add(72 * time.Hour)},
				Event: shared.BaseBillingEvent{
					Type: shared.EventTypeDirectDebitMandateCreated,
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 5,
				Date: shared.Date{Time: now.Add(180 * time.Hour)},
				Event: shared.BaseBillingEvent{
					Type: shared.EventTypeDirectDebitMandateCancelled,
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
	}

	assert.Equalf(t, expected, processPaymentMethodEvents(paymentMethods, 33), "processPaymentMethodsEvents(%v)", paymentMethods)
}

func Test_processDirectDebitEvents(t *testing.T) {
	tomorrow := time.Now().AddDate(0, 0, 1)
	now := time.Now()
	yesterday := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)
	aWeekAgo := time.Date(now.Year(), now.Month(), now.Day()-7, 0, 0, 0, 0, time.UTC)
	twoWeeksAgo := time.Date(now.Year(), now.Month(), now.Day()-14, 0, 0, 0, 0, time.UTC)

	directDebits := []store.GetDirectDebitPaymentsForBillingHistoryRow{
		{
			FinanceClientID: pgtype.Int4{Int32: 43},
			CollectionDate:  pgtype.Date{Time: tomorrow},
			Amount:          int32(123),
			Status:          "PENDING",
			LedgerID:        pgtype.Int4{Int32: 2},
			CreatedAt:       pgtype.Timestamp(pgtype.Date{Time: yesterday, Valid: true}),
			CreatedBy:       1,
		},
		{
			FinanceClientID: pgtype.Int4{Int32: 43},
			CollectionDate:  pgtype.Date{Time: yesterday, Valid: true},
			Amount:          int32(333),
			Status:          "COLLECTED",
			LedgerID:        pgtype.Int4{Int32: 3},
			CreatedAt:       pgtype.Timestamp(pgtype.Date{Time: aWeekAgo, Valid: true}),
			CreatedBy:       2,
		},
		{
			FinanceClientID: pgtype.Int4{Int32: 43},
			CollectionDate:  pgtype.Date{Time: yesterday},
			Amount:          int32(111),
			Status:          "FAILED",
			LedgerID:        pgtype.Int4{Int32: 4},
			CreatedAt:       pgtype.Timestamp(pgtype.Date{Time: twoWeeksAgo, Valid: true}),
			CreatedBy:       1,
		},
	}

	expected := []historyHolder{
		{
			billingHistory: shared.BillingHistory{
				User: 1,
				Date: shared.Date{Time: yesterday},
				Event: shared.DirectDebitEvent{
					Amount:         123,
					CollectionDate: shared.Date{Time: tomorrow},
					BaseBillingEvent: shared.BaseBillingEvent{
						Type: shared.EventTypeDirectDebitCollectionScheduled,
					},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 2,
				Date: shared.Date{Time: yesterday},
				Event: shared.DirectDebitEvent{
					Amount:         333,
					CollectionDate: shared.Date{Time: yesterday},
					BaseBillingEvent: shared.BaseBillingEvent{
						Type: shared.EventTypeDirectDebitCollected,
					},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 2,
				Date: shared.Date{Time: aWeekAgo},
				Event: shared.DirectDebitEvent{
					Amount:         333,
					CollectionDate: shared.Date{Time: yesterday},
					BaseBillingEvent: shared.BaseBillingEvent{
						Type: shared.EventTypeDirectDebitCollectionScheduled,
					},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 1,
				Date: shared.Date{Time: yesterday},
				Event: shared.DirectDebitEvent{
					Amount:         111,
					CollectionDate: shared.Date{Time: yesterday},
					BaseBillingEvent: shared.BaseBillingEvent{
						Type: shared.EventTypeDirectDebitCollectionFailed,
					},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
		{
			billingHistory: shared.BillingHistory{
				User: 1,
				Date: shared.Date{Time: twoWeeksAgo},
				Event: shared.DirectDebitEvent{
					Amount:         111,
					CollectionDate: shared.Date{Time: yesterday},
					BaseBillingEvent: shared.BaseBillingEvent{
						Type: shared.EventTypeDirectDebitCollectionScheduled,
					},
				},
				OutstandingBalance: 0,
			},
			balanceAdjustment: 0,
		},
	}

	assert.Equalf(t, expected, processDirectDebitEvents(directDebits), "processDirectDebitEvents(%v)", directDebits)
}
