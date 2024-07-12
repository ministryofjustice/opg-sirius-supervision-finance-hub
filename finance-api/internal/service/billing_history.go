package service

import (
	"context"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"sort"
	"strconv"
	"strings"
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
	user        string
	breakdown   []shared.PaymentBreakdown
	clientId    string
}

func (s *Service) GetBillingHistory(clientID int) ([]shared.BillingHistory, error) {
	ctx := context.Background()
	var history []historyHolder

	invoices, err := s.store.GetGeneratedInvoices(ctx, int32(clientID))
	if err != nil {
		return nil, err
	}

	history = invoiceEvents(invoices, history, strconv.Itoa(clientID))

	pendingAllocations, err := s.store.GetPendingLedgerAllocations(ctx, int32(clientID))
	if err != nil {
		return nil, err
	}

	allocationsByLedger := aggregateAllocations(pendingAllocations, strconv.Itoa(clientID))
	history = append(history, processAllocations(allocationsByLedger)...)
	sortHistoryByDate(history)

	return computeBillingHistory(history), nil
}

func invoiceEvents(invoices []store.GetGeneratedInvoicesRow, history []historyHolder, clientId string) []historyHolder {
	for _, inv := range invoices {
		bh := shared.BillingHistory{
			User: strconv.Itoa(int(inv.CreatedbyID.Int32)),
			Date: shared.Date{Time: inv.InvoiceDate.Time},
			Event: shared.InvoiceGenerated{
				ClientId: clientId,
				BaseBillingEvent: shared.BaseBillingEvent{
					Type: shared.EventTypeInvoiceGenerated,
				},
				InvoiceReference: shared.InvoiceEvent{
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
	return history
}

func aggregateAllocations(pendingAllocations []store.GetPendingLedgerAllocationsRow, clientID string) map[int32]allocationHolder {
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
				user:        strconv.Itoa(int(allo.CreatedbyID.Int32)),
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
					AdjustmentType:   strings.ToLower(allo.ledgerType),
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

func sortHistoryByDate(history []historyHolder) {
	sort.Slice(history, func(i, j int) bool {
		return history[i].billingHistory.Date.Time.Before(history[j].billingHistory.Date.Time)
	})
}

func computeBillingHistory(history []historyHolder) []shared.BillingHistory {
	var outstanding int
	var billingHistory []shared.BillingHistory
	for _, bh := range history {
		outstanding += bh.balanceAdjustment
		bh.billingHistory.OutstandingBalance = outstanding
		billingHistory = append(billingHistory, bh.billingHistory)
	}
	sort.Slice(billingHistory, func(i, j int) bool {
		if billingHistory[i].Date.Time.Equal(billingHistory[j].Date.Time) {
			return billingHistory[i].OutstandingBalance > billingHistory[j].OutstandingBalance
		}
		return billingHistory[i].Date.Time.After(billingHistory[j].Date.Time)
	})

	return billingHistory
}
