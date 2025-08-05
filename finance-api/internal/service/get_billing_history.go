package service

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"log/slog"
	"math"
	"slices"
	"sort"
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

	pendingAdjustments, err := s.store.GetPendingInvoiceAdjustments(ctx, clientID)
	if err != nil {
		return nil, err
	}

	history = append(history, processPendingAdjustments(pendingAdjustments, clientID)...)

	rejectedAdjustments, err := s.store.GetRejectedInvoiceAdjustments(ctx, clientID)
	if err != nil {
		return nil, err
	}

	history = append(history, processRejectedAdjustments(rejectedAdjustments, clientID)...)

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

	return computeBillingHistory(history), nil
}

func processRejectedAdjustments(adjustments []store.GetRejectedInvoiceAdjustmentsRow, clientID int32) []historyHolder {
	var history []historyHolder
	for _, adjustment := range adjustments {

		bh := shared.BillingHistory{
			User: int(adjustment.UpdatedBy.Int32),
			Date: shared.Date{Time: adjustment.UpdatedAt.Time},
		}
		bh.Event = shared.InvoiceAdjustmentRejected{
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
		}

		history = append(history, historyHolder{
			billingHistory:    bh,
			balanceAdjustment: 0,
		})
	}

	return history
}

func processPendingAdjustments(adjustments []store.GetPendingInvoiceAdjustmentsRow, clientID int32) []historyHolder {
	var history []historyHolder
	for _, adjustment := range adjustments {
		bh := shared.BillingHistory{
			User: int(adjustment.CreatedBy),
			Date: shared.Date{Time: adjustment.CreatedAt.Time},
		}
		bh.Event = shared.InvoiceAdjustmentPending{
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
		}

		history = append(history, historyHolder{
			billingHistory:    bh,
			balanceAdjustment: 0,
		})
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
				billingHistory:    bh,
				balanceAdjustment: 0,
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
			billingHistory:    bh,
			balanceAdjustment: 0,
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
			lh.billingHistory.Event = event
		} else {
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
				if allocation.Status == "ALLOCATED" {
					event.BaseBillingEvent = shared.BaseBillingEvent{
						Type: shared.EventTypePaymentProcessed,
					}
				}
			default:
				event.BaseBillingEvent = shared.BaseBillingEvent{
					Type: shared.EventTypeUnknown,
				}
			}

			// the allocated amounts should equal the total transaction for the event, excluding unapplies
			if allocation.Status != "UNAPPLIED" {
				event.Amount += int(allocation.AllocationAmount)
			}

			lh = &historyHolder{
				billingHistory: shared.BillingHistory{
					User:  int(allocation.CreatedBy.Int32),
					Date:  shared.Date{Time: allocation.LedgerDatetime.Time},
					Event: event,
				},
			}

			if event.TransactionType.IsPayment() && allocation.Status == "ALLOCATED" {
				lh.billingHistory.Date = shared.Date{Time: allocation.CreatedAt.Time}
			}
		}

		switch allocation.Status {
		case "ALLOCATED":
			lh.balanceAdjustment -= int(allocation.AllocationAmount)
		case "UNAPPLIED":
			lh.balanceAdjustment -= int(allocation.AllocationAmount)
			lh.creditAdjustment -= int(allocation.AllocationAmount)
		case "REAPPLIED":
			lh.balanceAdjustment -= int(allocation.AllocationAmount)
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

func computeBillingHistory(history []historyHolder) []shared.BillingHistory {
	// reverse order to allow for balance to be calculated
	sort.Slice(history, func(i, j int) bool {
		if history[i].billingHistory.Date.Time.Equal(history[j].billingHistory.Date.Time) {
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
