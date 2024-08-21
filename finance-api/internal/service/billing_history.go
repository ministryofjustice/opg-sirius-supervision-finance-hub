package service

import (
	"context"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"sort"
)

type historyHolder struct {
	billingHistory    shared.BillingHistory
	balanceAdjustment int
}

type allocationHolder struct {
	status      string
	ledgerType  string
	notes       string
	createdDate shared.Date
	user        int
	breakdown   []shared.PaymentBreakdown
	clientId    int
}

func (s *Service) GetBillingHistory(ctx context.Context, clientID int) ([]shared.BillingHistory, error) {
	invoices, err := s.store.GetGeneratedInvoices(ctx, int32(clientID))
	if err != nil {
		return nil, err
	}

	history := invoiceEvents(invoices, clientID)

	pendingAllocations, err := s.store.GetPendingLedgerAllocations(ctx, int32(clientID))
	if err != nil {
		return nil, err
	}

	allocationsByLedger := aggregateAllocations(pendingAllocations, clientID)
	history = append(history, processAllocations(allocationsByLedger)...)

	feEvents, err := s.store.GetFeeReductionEvents(ctx, int32(clientID))
	if err != nil {
		return nil, err
	}

	history = append(history, processFeeReductionEvents(feEvents)...)

	return computeBillingHistory(history), nil
}

func invoiceEvents(invoices []store.GetGeneratedInvoicesRow, clientID int) []historyHolder {
	var history []historyHolder
	for _, inv := range invoices {
		bh := shared.BillingHistory{
			User: int(inv.CreatedbyID.Int32),
			Date: shared.Date{Time: inv.InvoiceDate.Time},
			Event: shared.InvoiceGenerated{
				ClientId: clientID,
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

func aggregateAllocations(pendingAllocations []store.GetPendingLedgerAllocationsRow, clientID int) map[int32]allocationHolder {
	allocationsByLedger := make(map[int32]allocationHolder)
	for _, allo := range pendingAllocations {
		a, ok := allocationsByLedger[allo.LedgerID]
		if !ok {
			a = allocationHolder{
				status:      allo.Status,
				ledgerType:  allo.Type,
				notes:       allo.Notes.String,
				clientId:    clientID,
				createdDate: shared.Date{Time: allo.Datetime.Time},
				user:        int(allo.CreatedbyID.Int32),
				breakdown:   []shared.PaymentBreakdown{},
			}
		}
		a.breakdown = append(a.breakdown, shared.PaymentBreakdown{
			InvoiceReference: shared.InvoiceEvent{
				ID:        int(allo.InvoiceID),
				Reference: allo.Reference,
			},
			Amount: int(allo.Amount),
		})
		allocationsByLedger[allo.LedgerID] = a
	}
	return allocationsByLedger
}

func processAllocations(allocationsByLedger map[int32]allocationHolder) []historyHolder {
	var history []historyHolder
	for _, allo := range allocationsByLedger {
		bh := shared.BillingHistory{
			User: allo.user,
			Date: allo.createdDate,
		}

		if allo.status == "PENDING" {
			switch allo.ledgerType {
			case "CREDIT MEMO", "DEBIT MEMO", "CREDIT WRITE OFF":
				bh.Event = shared.InvoiceAdjustmentPending{
					BaseBillingEvent: shared.BaseBillingEvent{
						Type: shared.EventTypeInvoiceAdjustmentPending,
					},
					AdjustmentType:   shared.ParseAdjustmentType(allo.ledgerType),
					Notes:            allo.notes,
					PaymentBreakdown: allo.breakdown[0],
					ClientId:         allo.clientId,
				}
			}
		}

		history = append(history, historyHolder{
			billingHistory:    bh,
			balanceAdjustment: 0,
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
		if !fe.CancelledBy.Valid {
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
		if fe.Status.String == "APPROVED" {
			bh = shared.BillingHistory{
				User: int(fe.CreatedBy.Int32),
				Date: shared.Date{Time: fe.LedgerDate.Time},
				Event: shared.FeeReductionApplied{
					BaseBillingEvent: shared.BaseBillingEvent{
						Type: shared.EventTypeFeeReductionApplied,
					},
					ReductionType: shared.ParseFeeReductionType(fe.Type),
					PaymentBreakdown: shared.PaymentBreakdown{
						InvoiceReference: shared.InvoiceEvent{
							ID:        int(fe.InvoiceID.Int32),
							Reference: fe.Reference.String,
						},
						Amount: int(fe.Amount.Int32),
					},
					ClientId: int(fe.ClientID),
				},
			}
			history = append(history, historyHolder{
				billingHistory:    bh,
				balanceAdjustment: -(int(fe.Amount.Int32)),
			})
		}
	}
	return history
}

func computeBillingHistory(history []historyHolder) []shared.BillingHistory {
	// reverse order to allow for balance to be calculated
	sort.Slice(history, func(i, j int) bool {
		if history[i].billingHistory.Date.Time.Equal(history[j].billingHistory.Date.Time) {
			if _, ok := history[i].billingHistory.Event.(shared.FeeReductionApplied); ok {
				return history[i].billingHistory.OutstandingBalance < history[j].billingHistory.OutstandingBalance
			}
			return history[i].billingHistory.OutstandingBalance > history[j].billingHistory.OutstandingBalance
		}
		return history[i].billingHistory.Date.Time.Before(history[j].billingHistory.Date.Time)
	})

	var outstanding int
	var billingHistory []shared.BillingHistory
	for _, bh := range history {
		outstanding += bh.balanceAdjustment
		bh.billingHistory.OutstandingBalance = outstanding
		billingHistory = append(billingHistory, bh.billingHistory)
	}

	// flip it back
	sort.Slice(billingHistory, func(i, j int) bool {
		if billingHistory[i].Date.Time.Equal(billingHistory[j].Date.Time) {
			if _, ok := billingHistory[i].Event.(shared.FeeReductionApplied); ok {
				return billingHistory[i].OutstandingBalance < billingHistory[j].OutstandingBalance
			}
			return billingHistory[i].OutstandingBalance > billingHistory[j].OutstandingBalance
		}
		return billingHistory[i].Date.Time.After(billingHistory[j].Date.Time)
	})

	return billingHistory
}
