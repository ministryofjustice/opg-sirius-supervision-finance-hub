package service

import (
	"context"
	"github.com/opg-sirius-finance-hub/shared"
	"sort"
	"strconv"
)

// holds the value by which to increment/decrement the outstanding balance
type historyHolder struct {
	billingHistory    shared.BillingHistory
	balanceAdjustment int
}

// allows flat rows to be mapped per ledger
type allocationHolder struct {
	ledgerType  string
	notes       string
	createdDate shared.Date
	user        string
	breakdown   []shared.PaymentBreakdown
}

func (s *Service) GetBillingHistory(clientID int) ([]shared.BillingHistory, error) {
	ctx := context.Background()

	// query each table:
	// invoice generated
	invoices, err := s.store.GetGeneratedInvoices(ctx, int32(clientID))
	if err != nil {
		return nil, err
	}

	var history []historyHolder

	for _, inv := range invoices {
		bh := shared.BillingHistory{
			User: strconv.Itoa(int(inv.CreatedbyID.Int32)), // need assignees table access
			Date: shared.Date{Time: inv.Createddate.Time},
			Event: shared.InvoiceGenerated{
				BaseBillingEvent: shared.BaseBillingEvent{
					Type: shared.EventTypeInvoiceGenerated,
				},
				InvoiceReference: shared.InvoiceReference{
					ID:        int(inv.InvoiceID),
					Reference: inv.Reference,
				},
				InvoiceType: inv.Feetype,
				Amount:      int(inv.Amount),
			},
		}

		history = append(history, historyHolder{
			billingHistory:    bh,
			balanceAdjustment: int(inv.Amount),
		})
	}

	// fetch all applied ledger allocations (adjustments, reductions, payments)
	appliedAllocations, err := s.store.GetAppliedLedgerAllocations(ctx, int32(clientID))
	if err != nil {
		return nil, err
	}

	// a ledger can include allocations to multiple invoices, so first we need to group them by ledger
	allocationsByLedger := make(map[int32]allocationHolder)
	for _, allo := range appliedAllocations {
		a, ok := allocationsByLedger[allo.LedgerID]
		if !ok {
			a = allocationHolder{
				ledgerType:  allo.Type,
				notes:       allo.Notes.String,
				createdDate: shared.Date{Time: allo.Createddate.Time},
				user:        strconv.Itoa(int(allo.CreatedbyID.Int32)),
				breakdown:   []shared.PaymentBreakdown{},
			}
		}
		a.breakdown = append(a.breakdown, shared.PaymentBreakdown{
			InvoiceReference: shared.InvoiceReference{
				ID:        int(allo.InvoiceID),
				Reference: allo.Reference,
			},
			Amount: int(allo.Amount),
		})
		allocationsByLedger[allo.LedgerID] = a
	}

	// now range through the row arrays to compile the events
	for _, allo := range allocationsByLedger {
		var amount int
		for _, breakdown := range allo.breakdown {
			amount += breakdown.Amount
		}

		bh := shared.BillingHistory{
			User: allo.user, // need assignees table access
			Date: allo.createdDate,
		}

		switch allo.ledgerType {
		case "CREDIT MEMO", "DEBIT MEMO", "CREDIT WRITE OFF":
			bh.Event = shared.InvoiceAdjustmentApproved{
				BaseBillingEvent: shared.BaseBillingEvent{
					Type: shared.EventTypeInvoiceAdjustmentApproved,
				},
				AdjustmentType:   allo.ledgerType,
				PaymentBreakdown: allo.breakdown[0], // adjustments apply to a single invoice
			}
		case "CREDIT EXEMPTION", "CREDIT HARDSHIP", "CREDIT REMISSION":
			bh.Event = shared.FeeReductionApplied{
				BaseBillingEvent: shared.BaseBillingEvent{
					Type: shared.EventTypeFeeReductionApplied,
				},
				ReductionType:    allo.ledgerType,   // could combine with above?
				PaymentBreakdown: allo.breakdown[0], // adjustments apply to a single invoice
			}
		case "BACS TRANSFER", "CARD PAYMENT", "UNKNOWN CREDIT", "UNKNOWN DEBIT": // check types
			bh.Event = shared.PaymentProcessed{
				BaseBillingEvent: shared.BaseBillingEvent{
					Type: shared.EventTypePaymentProcessed,
				},
				PaymentType: allo.ledgerType,
				Breakdown:   allo.breakdown,
				Total:       amount,
			}
		}

		history = append(history, historyHolder{
			billingHistory:    bh,
			balanceAdjustment: -amount, // allocated funds subtract from the outstanding balance
		})
	}

	// oldest first
	sort.Slice(history, func(i, j int) bool {
		return history[i].billingHistory.Date.Time.After(history[j].billingHistory.Date.Time)
	})

	// calculate balances by iterating through (approved) ledgers and invoices, then extract history from holder
	var outstanding int
	var billingHistory []shared.BillingHistory
	for _, bh := range history {
		outstanding += bh.balanceAdjustment
		bh.billingHistory.OutstandingBalance = outstanding

		billingHistory = append(billingHistory, bh.billingHistory)
	}

	// sort in the correct order
	sort.Slice(history, func(i, j int) bool {
		return history[i].billingHistory.Date.Time.Before(history[j].billingHistory.Date.Time)
	})

	return billingHistory, nil
}
