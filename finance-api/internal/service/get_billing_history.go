package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"sort"
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type historyHolder struct {
	billingHistory    shared.BillingHistory
	balanceAdjustment int
	creditAdjustment  int
}

func (s *Service) GetBillingHistory(ctx context.Context, clientID int32) ([]shared.BillingHistory, error) {
	invoices, err := s.store.GetGeneratedInvoices(ctx, clientID)

	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Error in getting invoices in billing history for client %d", clientID), slog.String("err", err.Error()))
		return nil, err
	}

	history := invoiceEvents(invoices, clientID)

	adjustments, err := s.store.GetInvoiceAdjustmentEvents(ctx, clientID)
	if err != nil {
		return nil, err
	}

	history = append(history, processAdjustments(adjustments, clientID)...)

	feEvents, err := s.store.GetFeeReductionEvents(ctx, clientID)
	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Error in getting fee reductions events in billing history for client %d", clientID), slog.String("err", err.Error()))
		return nil, err
	}

	history = append(history, processFeeReductionEvents(feEvents)...)

	allocations, err := s.store.GetLedgerAllocationsForClient(ctx, clientID)
	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Error in getting ledger allocations in billing history for client %d", clientID), slog.String("err", err.Error()))
		return nil, err
	}

	history = append(history, processLedgerAllocations(allocations, clientID)...)

	refunds, err := s.store.GetRefundsForBillingHistory(ctx, clientID)
	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Error in getting refunds in billing history for client %d", clientID), slog.String("err", err.Error()))
		return nil, err
	}

	history = append(history, processRefundEvents(refunds, clientID)...)

	paymentMethods, err := s.store.GetPaymentMethodsForBillingHistory(ctx, clientID)
	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Error in getting payment methods in billing history for client %d", clientID), slog.String("err", err.Error()))
		return nil, err
	}

	history = append(history, processPaymentMethodEvents(paymentMethods)...)

	directDebitEvents, err := s.store.GetDirectDebitPaymentsForBillingHistory(ctx, clientID)
	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Error in getting Direct Debit payments in billing history for client %d", clientID), slog.String("err", err.Error()))
		return nil, err
	}

	history = append(history, processDirectDebitEvents(directDebitEvents)...)

	return computeBillingHistory(history), nil
}

func processAdjustments(adjustments []store.GetInvoiceAdjustmentEventsRow, clientID int32) []historyHolder {
	var history []historyHolder
	for _, adjustment := range adjustments {
		// every adjustment will have a pending event on creation
		pending := shared.BillingHistory{
			User: int(adjustment.CreatedBy),
			Date: shared.Date{Time: adjustment.CreatedAt.Time},
			Event: shared.InvoiceAdjustmentPending{
				BaseBillingEvent: shared.BaseBillingEvent{
					Type: shared.EventTypeInvoiceAdjustmentPending,
				},
				AdjustmentType: shared.ParseAdjustmentType(adjustment.AdjustmentType),
				Notes:          adjustment.Notes,
				ClientId:       int(clientID),
				PaymentBreakdown: shared.PaymentBreakdown{
					InvoiceReference: shared.InvoiceEvent{
						ID:        int(adjustment.InvoiceID),
						Reference: adjustment.Reference,
					},
					Amount: int(adjustment.Amount),
				},
			},
		}

		history = append(history, historyHolder{
			billingHistory: pending,
		})

		if adjustment.Status == "REJECTED" {
			rejected := shared.BillingHistory{
				User: int(adjustment.UpdatedBy.Int32),
				Date: shared.Date{Time: adjustment.UpdatedAt.Time},
				Event: shared.InvoiceAdjustmentRejected{
					BaseBillingEvent: shared.BaseBillingEvent{
						Type: shared.EventTypeInvoiceAdjustmentRejected,
					},
					AdjustmentType: shared.ParseAdjustmentType(adjustment.AdjustmentType),
					Notes:          adjustment.Notes,
					ClientId:       int(clientID),
					PaymentBreakdown: shared.PaymentBreakdown{
						InvoiceReference: shared.InvoiceEvent{
							ID:        int(adjustment.InvoiceID),
							Reference: adjustment.Reference,
						},
						Amount: int(adjustment.Amount),
					},
				},
			}
			history = append(history, historyHolder{
				billingHistory: rejected,
			})
		}
	}

	return history
}

func invoiceEvents(invoices []store.GetGeneratedInvoicesRow, clientID int32) []historyHolder {
	var history []historyHolder
	for _, inv := range invoices {
		bh := shared.BillingHistory{
			User: int(inv.CreatedBy.Int32),
			Date: shared.Date{Time: inv.CreatedAt.Time},
			Event: shared.InvoiceGenerated{
				ClientId: int(clientID),
				BaseBillingEvent: shared.BaseBillingEvent{
					Type: shared.EventTypeInvoiceGenerated,
				},
				InvoiceReference: shared.InvoiceEvent{
					ID:        int(inv.InvoiceID),
					Reference: inv.Reference,
				},
				InvoiceType: shared.ParseInvoiceType(inv.Feetype),
				Amount:      int(inv.Amount),
			},
		}

		history = append(history, historyHolder{
			billingHistory:    bh,
			balanceAdjustment: int(inv.Amount),
		})
	}
	return history
}

func getRefundEventTypeAndDate(refund store.GetRefundsForBillingHistoryRow) (shared.BillingEventType, time.Time) {
	if refund.CancelledAt.Valid {
		return shared.EventTypeRefundCancelled, refund.CancelledAt.Time
	} else if refund.Decision == "REJECTED" {
		return shared.EventTypeRefundStatusUpdated, refund.DecisionAt.Time
	} else if refund.ProcessedAt.Valid {
		//	approved status with decision at has a date and processed at is set makes processing event
		return shared.EventTypeRefundProcessing, refund.ProcessedAt.Time
	} else if refund.Decision == "APPROVED" {
		//	approved status with decision at has a date and processed at is null
		return shared.EventTypeRefundApproved, refund.DecisionAt.Time
	}
	return shared.EventTypeRefundCreated, refund.CreatedAt.Time
}

func getUserForEventType(refund store.GetRefundsForBillingHistoryRow, eventType shared.BillingEventType) int32 {
	switch eventType {
	case shared.EventTypeRefundCreated:
		return refund.CreatedBy
	case shared.EventTypeRefundCancelled:
		return refund.CancelledBy.Int32
	default:
		return refund.DecisionBy.Int32
	}
}

func makeRefundEvent(refund store.GetRefundsForBillingHistoryRow, user int32, eventType shared.BillingEventType, date time.Time, clientID int32, history []historyHolder) []historyHolder {
	bh := shared.BillingHistory{
		User: int(user),
		Date: shared.Date{Time: date},
		Event: shared.RefundEvent{
			Id:               int(refund.RefundID),
			ClientId:         int(clientID),
			Amount:           int(refund.Amount),
			BaseBillingEvent: shared.BaseBillingEvent{Type: eventType},
			Notes:            refund.Notes,
		},
	}
	//never change balance as it will double count due to the ledger billing history event
	history = append(history, historyHolder{
		billingHistory: bh,
	})
	return history
}

func processRefundEvents(refunds []store.GetRefundsForBillingHistoryRow, clientID int32) []historyHolder {
	var history []historyHolder
	for _, re := range refunds {
		// for all refunds ensure that there is a refund created event - this is done in the final makeRefundEvent
		if re.Decision != "PENDING" {
			eventType, date := getRefundEventTypeAndDate(re)
			user := getUserForEventType(re, eventType)
			history = makeRefundEvent(re, user, eventType, date, clientID, history)

			if eventType == shared.EventTypeRefundCancelled {
				//	ensure there is a second timeline event for the approved and processing events
				if re.ProcessedAt.Valid {
					history = makeRefundEvent(re, re.DecisionBy.Int32, shared.EventTypeRefundProcessing, re.ProcessedAt.Time, clientID, history)
				}
				history = makeRefundEvent(re, re.DecisionBy.Int32, shared.EventTypeRefundApproved, re.DecisionAt.Time, clientID, history)
			}

			if eventType == shared.EventTypeRefundProcessing {
				history = makeRefundEvent(re, re.DecisionBy.Int32, shared.EventTypeRefundApproved, re.DecisionAt.Time, clientID, history)
			}
		}

		history = makeRefundEvent(re, re.CreatedBy, shared.EventTypeRefundCreated, re.CreatedAt.Time, clientID, history)
	}

	return history
}

func processDirectDebitEvents(directDebitEvents []store.GetDirectDebitPaymentsForBillingHistoryRow) []historyHolder {
	var history []historyHolder
	for _, dd := range directDebitEvents {
		bh := shared.BillingHistory{
			User: int(dd.CreatedBy),
			Date: shared.Date{Time: dd.CreatedAt.Time},
			Event: shared.DirectDebitEvent{
				Amount:           int(dd.Amount),
				CollectionDate:   shared.Date{Time: dd.CollectionDate.Time},
				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeDirectDebitCollectionScheduled},
			},
		}

		// balance is only adjusted by ledgers, not Direct Debit events
		history = append(history, historyHolder{
			billingHistory: bh,
		})
	}
	return history
}

func processFeeReductionEvents(feEvents []store.GetFeeReductionEventsRow) []historyHolder {
	var history []historyHolder
	for _, fe := range feEvents {
		var bh shared.BillingHistory
		if fe.CancelledBy.Valid {
			bh = shared.BillingHistory{
				User: int(fe.CancelledBy.Int32),
				Date: shared.Date{Time: fe.CancelledAt.Time},
				Event: shared.FeeReductionCancelled{
					ReductionType:      shared.ParseFeeReductionType(fe.Type),
					CancellationReason: fe.CancellationReason.String,
					BaseBillingEvent: shared.BaseBillingEvent{
						Type: shared.EventTypeFeeReductionCancelled,
					},
				},
			}
			history = append(history, historyHolder{
				billingHistory: bh,
			})
		}
		bh = shared.BillingHistory{
			User: int(fe.CreatedBy.Int32),
			Date: shared.Date{Time: fe.CreatedAt.Time},
			Event: shared.FeeReductionAwarded{
				ReductionType: shared.ParseFeeReductionType(fe.Type),
				StartDate:     shared.Date{Time: fe.Startdate.Time},
				EndDate:       shared.Date{Time: fe.Enddate.Time},
				DateReceived:  shared.Date{Time: fe.Datereceived.Time},
				Notes:         fe.Notes,
				BaseBillingEvent: shared.BaseBillingEvent{
					Type: shared.EventTypeFeeReductionAwarded,
				},
			},
		}
		history = append(history, historyHolder{
			billingHistory: bh,
		})
	}
	return history
}

// processLedgerAllocations takes an array of allocations and groups them by ledger, which defines a single billing event.
// A ledger is always for a single transaction type but may have multiple allocations associated with it.
func processLedgerAllocations(allocations []store.GetLedgerAllocationsForClientRow, clientID int32) []historyHolder {
	historyByLedger := make(map[int32]*historyHolder)

	for _, allocation := range allocations {
		var (
			event shared.TransactionEvent
			lh    *historyHolder
			ok    bool
		)
		if lh, ok = historyByLedger[allocation.LedgerID]; ok {
			// there will only be one key transaction type per ledger, so add transaction to payment breakdown
			event = lh.billingHistory.Event.(shared.TransactionEvent)
			event.Breakdown = append(event.Breakdown,
				shared.PaymentBreakdown{InvoiceReference: shared.InvoiceEvent{
					ID:        int(allocation.InvoiceID.Int32),
					Reference: allocation.Reference.String,
				},
					Amount: int(math.Abs(float64(allocation.AllocationAmount))),
					Status: allocation.Status,
				},
			)
			if !event.TransactionType.IsReceiptTransaction() {
				event.Amount += int(math.Abs(float64(allocation.AllocationAmount)))
			}
			lh.billingHistory.Event = event
		} else {
			lh = &historyHolder{
				billingHistory: shared.BillingHistory{
					User: int(allocation.CreatedBy.Int32),
					Date: shared.Date{Time: allocation.LedgerDatetime.Time},
				},
			}

			event = shared.TransactionEvent{
				ClientId:        int(clientID),
				TransactionType: shared.ParseTransactionType(allocation.Type),
				Breakdown: []shared.PaymentBreakdown{
					{
						InvoiceReference: shared.InvoiceEvent{
							ID:        int(allocation.InvoiceID.Int32),
							Reference: allocation.Reference.String,
						},
						Amount: int(math.Abs(float64(allocation.AllocationAmount))),
						Status: allocation.Status,
					},
				},
				Amount:           int(allocation.LedgerAmount),
				BaseBillingEvent: shared.BaseBillingEvent{},
			}
			switch {
			case shared.ParseFeeReductionType(allocation.Type).Valid():
				event.BaseBillingEvent = shared.BaseBillingEvent{
					Type: shared.EventTypeFeeReductionApplied,
				}
			case shared.ParseAdjustmentType(allocation.Type).Valid():
				event.BaseBillingEvent = shared.BaseBillingEvent{
					Type: shared.EventTypeInvoiceAdjustmentApplied,
				}
			case event.TransactionType == shared.TransactionTypeReapply:
				event.BaseBillingEvent = shared.BaseBillingEvent{
					Type: shared.EventTypeReappliedCredit,
				}
			case event.TransactionType.IsPayment():
				event.BaseBillingEvent = shared.BaseBillingEvent{
					Type: shared.EventTypePaymentProcessed,
				}
				lh.billingHistory.Date = shared.Date{Time: allocation.CreatedAt.Time}
			case event.TransactionType == shared.TransactionTypeRefund:
				event.BaseBillingEvent = shared.BaseBillingEvent{
					Type: shared.EventTypeRefundProcessed,
				}
				lh.billingHistory.Date = shared.Date{Time: allocation.CreatedAt.Time}
			default:
				event.BaseBillingEvent = shared.BaseBillingEvent{
					Type: shared.EventTypeUnknown,
				}
			}

			lh.billingHistory.Event = event
		}

		// receipt transactions that don't apply to an invoice will only affect credit
		if !event.TransactionType.IsReceiptTransaction() || allocation.InvoiceID.Valid {
			lh.balanceAdjustment -= int(allocation.AllocationAmount)
		}

		if allocation.Status == "UNAPPLIED" || allocation.Status == "REAPPLIED" {
			lh.creditAdjustment -= int(allocation.AllocationAmount)
		}

		historyByLedger[allocation.LedgerID] = lh
	}

	var history []historyHolder
	for _, lh := range historyByLedger {
		history = append(history, *lh)
	}

	return history
}

func processPaymentMethodEvents(paymentMethods []store.GetPaymentMethodsForBillingHistoryRow) []historyHolder {
	var history []historyHolder
	for _, pm := range paymentMethods {
		var eventType shared.BillingEventType
		if pm.Type == "DEMANDED" {
			eventType = shared.EventTypeDirectDebitMandateCancelled
		} else {
			eventType = shared.EventTypeDirectDebitMandateCreated
		}
		bh := shared.BillingHistory{
			User: int(pm.CreatedBy),
			Date: shared.Date{Time: pm.CreatedAt.Time},
			Event: shared.PaymentMethodChangedEvent{
				BaseBillingEvent: shared.BaseBillingEvent{
					Type: eventType,
				},
			},
		}
		history = append(history, historyHolder{
			billingHistory: bh,
		})
	}

	return history
}

func computeBillingHistory(history []historyHolder) []shared.BillingHistory {
	// reverse order to allow for balance to be calculated
	sort.Slice(history, func(i, j int) bool {
		if history[i].billingHistory.Date.Time.Equal(history[j].billingHistory.Date.Time) {
			// Direct Debit mandate created should always appear before the schedule creation
			if history[i].billingHistory.Event.GetType() == shared.EventTypeDirectDebitCollectionScheduled {
				return history[j].billingHistory.Event.GetType() != shared.EventTypeDirectDebitMandateCreated
			}

			// reapplies should apply after if they are the result of a transaction event
			if _, ok := history[i].billingHistory.Event.(shared.TransactionEvent); ok {
				return history[j].billingHistory.Event.GetType() == shared.EventTypeReappliedCredit
			}

			// transaction events and reapplies should apply after the event that causes them
			return history[i].billingHistory.Event.GetType() != shared.EventTypeReappliedCredit
		}
		return history[i].billingHistory.Date.Time.Before(history[j].billingHistory.Date.Time)
	})

	var (
		outstanding    int
		credit         int
		billingHistory []shared.BillingHistory
	)
	for _, bh := range history {
		outstanding += bh.balanceAdjustment
		credit += bh.creditAdjustment
		bh.billingHistory.OutstandingBalance = outstanding
		bh.billingHistory.CreditBalance = credit
		billingHistory = append(billingHistory, bh.billingHistory)
	}

	// flip it back
	slices.Reverse(billingHistory)

	return billingHistory
}
