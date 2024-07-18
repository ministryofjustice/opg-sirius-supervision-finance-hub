package service

import (
	"context"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"math"
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

func (s *Service) GetBillingHistory(ctx context.Context, clientId int) ([]shared.BillingHistory, error) {
	pendingAllocations, err := s.store.GetPendingLedgerAllocations(ctx, int32(clientId))
	if err != nil {
		return nil, err
	}

	allocationsByLedger := aggregateAllocations(pendingAllocations, strconv.Itoa(clientId))

	startingBalance, err := s.store.GetAccountInformation(ctx, int32(clientId))
	if err != nil {
		return nil, err
	}
	history := processPendingAllocations(allocationsByLedger, int(startingBalance.Outstanding))

	approvedAllocations, err := s.store.GetApprovedLedgerAllocations(ctx, int32(clientId))
	if err != nil {
		return nil, err
	}

	allocationsByLedger = aggregateApprovedAllocations(approvedAllocations, strconv.Itoa(clientId))
	approvedAllocationsHistory := processApprovedAllocations(allocationsByLedger, int(startingBalance.Outstanding))

	history = append(history, approvedAllocationsHistory...)

	sortHistoryByDate(history)
	return computeBillingHistory(history), nil
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
				createdDate: shared.Date{Time: allo.Createddate.Time},
				user:        strconv.Itoa(int(allo.CreatedbyID.Int32)),
				breakdown:   []shared.PaymentBreakdown{},
			}
		}
		a.breakdown = append(a.breakdown, shared.PaymentBreakdown{
			InvoiceReference: shared.InvoiceEvent{
				ID:        int(allo.InvoiceID),
				Reference: allo.Reference,
			},
			Amount: int(math.Abs(float64(allo.Amount / 100))),
		})
		allocationsByLedger[allo.LedgerID] = a
	}
	return allocationsByLedger
}

func aggregateApprovedAllocations(pendingAllocations []store.GetApprovedLedgerAllocationsRow, clientID string) map[int32]allocationHolder {
	allocationsByLedger := make(map[int32]allocationHolder)
	for _, allo := range pendingAllocations {
		a, ok := allocationsByLedger[allo.LedgerID]
		if !ok {
			a = allocationHolder{
				status:      allo.Status,
				ledgerType:  allo.Type,
				notes:       allo.Notes.String,
				clientId:    clientID,
				createdDate: shared.Date{Time: allo.Createddate.Time},
				user:        strconv.Itoa(int(allo.CreatedbyID.Int32)),
				breakdown:   []shared.PaymentBreakdown{},
			}
		}
		a.breakdown = append(a.breakdown, shared.PaymentBreakdown{
			InvoiceReference: shared.InvoiceEvent{
				ID:        int(allo.InvoiceID),
				Reference: allo.Reference,
			},
			Amount: int(math.Abs(float64(allo.Amount / 100))),
		})
		allocationsByLedger[allo.LedgerID] = a
	}
	return allocationsByLedger
}

func processApprovedAllocations(allocationsByLedger map[int32]allocationHolder, startingBalance int) []historyHolder {
	var history []historyHolder
	amount := startingBalance / 100

	for _, allo := range allocationsByLedger {
		bh := shared.BillingHistory{
			User: allo.user,
			Date: allo.createdDate,
		}

		switch allo.ledgerType {
		case "HARDSHIP", "REMISSION", "EXEMPTION":
			bh.Event = shared.FeeReductionApplied{
				BaseBillingEvent: shared.BaseBillingEvent{
					Type: shared.EventTypeFeeReductionApplied,
				},
				ReductionType:    cases.Title(language.English).String(allo.ledgerType),
				PaymentBreakdown: allo.breakdown[0],
				ClientId:         allo.clientId,
			}
		}

		history = append(history, historyHolder{
			billingHistory:    bh,
			balanceAdjustment: amount,
		})
	}
	return history
}

func processPendingAllocations(allocationsByLedger map[int32]allocationHolder, startingBalance int) []historyHolder {
	var history []historyHolder
	amount := startingBalance / 100

	for _, allo := range allocationsByLedger {
		bh := shared.BillingHistory{
			User: allo.user,
			Date: allo.createdDate,
		}

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

		history = append(history, historyHolder{
			billingHistory:    bh,
			balanceAdjustment: amount,
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
		if bh.billingHistory.Event.GetType() == shared.EventTypeInvoiceAdjustmentPending {
			outstanding = bh.balanceAdjustment
			bh.billingHistory.OutstandingBalance = outstanding
			billingHistory = append(billingHistory, bh.billingHistory)
		} else {
			outstanding += bh.balanceAdjustment
			bh.billingHistory.OutstandingBalance = outstanding
			billingHistory = append(billingHistory, bh.billingHistory)
		}
	}
	sort.Slice(billingHistory, func(i, j int) bool {
		return billingHistory[i].Date.Time.After(billingHistory[j].Date.Time)
	})
	return billingHistory
}
