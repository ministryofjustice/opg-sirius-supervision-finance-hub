package service

import (
	"context"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"sort"
	"strconv"
	"strings"
)

type historyHolder struct {
	billingHistory    shared.BillingHistory
	balanceAdjustment int
}

type allocationHolder struct {
	status           string
	ledgerType       string
	feeReductionType string
	notes            string
	createdDate      shared.Date
	user             string
	breakdown        []shared.PaymentBreakdown
	clientId         string
}

func (s *Service) GetBillingHistory(ctx context.Context, clientID int) ([]shared.BillingHistory, error) {
	var history []historyHolder

	invoices, err := s.store.GetGeneratedInvoices(ctx, int32(clientID))
	if err != nil {
		return nil, err
	}

	history = invoiceEvents(invoices, history, strconv.Itoa(clientID))

	allocations, err := s.store.GetClientLedgerAllocations(ctx, int32(clientID))
	if err != nil {
		return nil, err
	}

	history = allocationEvents(allocations, history, strconv.Itoa(clientID))

	sortHistoryByDate(history)
	return computeBillingHistory(history), nil
}

func allocationEvents(allocations []store.GetClientLedgerAllocationsRow, history []historyHolder, clientId string) []historyHolder {
	allocationsByLedger := aggregateAllocations(allocations, clientId)

	var balanceAdjustment int
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
				balanceAdjustment = 0
			}
		}

		if allo.status == "APPROVED" {
			switch allo.feeReductionType {
			case "HARDSHIP", "REMISSION", "EXEMPTION":
				bh.Event = shared.FeeReductionApplied{
					BaseBillingEvent: shared.BaseBillingEvent{
						Type: shared.EventTypeFeeReductionApplied,
					},
					ReductionType:    cases.Title(language.English).String(allo.feeReductionType),
					PaymentBreakdown: allo.breakdown[0],
					ClientId:         allo.clientId,
				}
				balanceAdjustment = -(allo.breakdown[0].Amount)
			}
		}

		history = append(history, historyHolder{
			billingHistory:    bh,
			balanceAdjustment: balanceAdjustment,
		})
	}
	return history
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

func aggregateAllocations(allocations []store.GetClientLedgerAllocationsRow, clientID string) map[int32]allocationHolder {
	allocationsByLedger := make(map[int32]allocationHolder)
	for _, allo := range allocations {
		a, ok := allocationsByLedger[allo.LedgerID]
		if !ok {
			a = allocationHolder{
				status:           allo.Status,
				ledgerType:       allo.LedgerType,
				feeReductionType: (allo.FeeReductionType).(string),
				notes:            allo.Notes.String,
				clientId:         clientID,
				createdDate:      shared.Date{Time: allo.Datetime.Time},
				user:             strconv.Itoa(int(allo.CreatedbyID.Int32)),
				breakdown:        []shared.PaymentBreakdown{},
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
