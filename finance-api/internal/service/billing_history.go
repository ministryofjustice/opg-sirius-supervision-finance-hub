package service

import (
	"context"
	"github.com/opg-sirius-finance-hub/shared"
	"sort"
	"strconv"
)

var (
	EventTypeInvoiceGenerated         = "EventTypeInvoiceGenerated"
	EventTypeFeeReductionApplied      = "EventTypeFeeReductionApplied"
	EventTypeInvoiceAdjustmentApplied = "EventTypeInvoiceAdjustmentApplied"
)

func (s *Service) GetBillingHistory(clientID int) ([]shared.BillingHistory, error) {
	ctx := context.Background()

	// query each table:
	// invoice generated
	invoices, err := s.store.GetGeneratedInvoices(ctx, int32(clientID))
	if err != nil {
		return nil, err
	}
	// added invoice adjustment (inc applied fee reductions)
	addedAdjustments, err := s.store.GetAddedInvoiceAdjustments(ctx, int32(clientID))
	if err != nil {
		return nil, err
	}

	var billingHistory []shared.BillingHistory

	for _, inv := range invoices {
		bh := shared.BillingHistory{
			InvoiceReference:   inv.Reference,
			User:               strconv.Itoa(int(inv.CreatedbyID.Int32)), // need assignees table access
			Date:               shared.Date{Time: inv.Createddate.Time},
			EventType:          EventTypeInvoiceGenerated,
			Type:               inv.Feetype,
			Amount:             int(inv.Amount),
			OutstandingBalance: 0, // not sure how to calculate this
		}

		billingHistory = append(billingHistory, bh)
	}

	for _, adj := range addedAdjustments {
		bh := shared.BillingHistory{
			InvoiceReference:   adj.Reference,
			User:               strconv.Itoa(int(adj.CreatedbyID.Int32)), // need assignees table access
			Date:               shared.Date{Time: adj.Createddate.Time},
			Type:               EventTypeInvoiceAdjustmentApplied,
			Amount:             int(adj.Amount),
			OutstandingBalance: 0, // not sure how to calculate this
		}

		if adj.FeeReductionType.String != "" {
			bh.Type = adj.FeeReductionType.String
			bh.EventType = EventTypeFeeReductionApplied
		}
		billingHistory = append(billingHistory, bh)
	}

	sort.Slice(billingHistory, func(i, j int) bool {
		return billingHistory[i].Date.Time.Before(billingHistory[j].Date.Time)
	})

	// calculate balances by iterating through (approved) ledgers and invoices

	return billingHistory, nil
}
