package service

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"golang.org/x/exp/maps"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"slices"
	"sort"
)

var (
	PaymentTypes = []string{"CARD PAYMENT", "BACS TRANSFER"}
)

type invoiceMetadata struct {
	invoice          *shared.Invoice
	contextType      string
	paymentReceived  bool
	feeReductionType string
}

type invoiceBuilder struct {
	invoices            map[int]*invoiceMetadata
	invoicePositionByID map[int32]int
}

func newInvoiceBuilder(invoices []store.GetInvoicesRow) *invoiceBuilder {
	ib := invoiceBuilder{make(map[int]*invoiceMetadata), make(map[int32]int)}
	for index, inv := range invoices {
		ib.invoices[index] = &invoiceMetadata{
			invoice: &shared.Invoice{
				Id:                 int(inv.ID),
				Ref:                inv.Reference,
				Amount:             int(inv.Amount),
				RaisedDate:         shared.Date{Time: inv.Raiseddate.Time},
				Status:             "Unpaid",
				Received:           int(inv.Received),
				OutstandingBalance: int(inv.Amount) - int(inv.Received),
			},
			feeReductionType: inv.FeeReductionType,
		}
		ib.invoicePositionByID[inv.ID] = index
	}
	return &ib
}

func (ib *invoiceBuilder) GetIDs() []int32 {
	return maps.Keys(ib.invoicePositionByID)
}

func (ib *invoiceBuilder) Build() shared.Invoices {
	var invoices shared.Invoices

	keys := maps.Keys(ib.invoices)
	sort.Ints(keys)

	for _, key := range keys {
		inv := ib.invoices[key]
		invoice := inv.invoice
		var status string
		if invoice.OutstandingBalance > 0 {
			status = "Unpaid"
		} else if invoice.OutstandingBalance < 0 {
			status = "Overpaid"
		} else if invoice.OutstandingBalance == 0 {
			if inv.paymentReceived {
				status = "Paid"
			} else {
				status = "Closed"
			}
		}

		if inv.feeReductionType != "" && inv.contextType == "" {
			inv.contextType = cases.Title(language.English).String(inv.feeReductionType)
		}

		if inv.contextType != "" {
			status = fmt.Sprintf("%s - %s", status, inv.contextType)
		}

		invoice.Status = status
		invoices = append(invoices, *inv.invoice)
	}
	return invoices
}

func (ib *invoiceBuilder) addLedgerAllocations(ilas []store.GetLedgerAllocationsRow) {
	writeOffReversed := false
	for _, il := range ilas {
		metadata := ib.invoices[ib.invoicePositionByID[il.InvoiceID.Int32]]

		metadata.invoice.Ledgers = append(
			metadata.invoice.Ledgers,
			shared.Ledger{
				Amount:          int(il.Amount),
				ReceivedDate:    shared.Date{Time: il.RaisedDate.Time},
				TransactionType: il.Type,
				Status:          il.Status,
			})

		if il.Type == "WRITE OFF REVERSAL" {
			writeOffReversed = true
		}
		if il.Status == "ALLOCATED" {
			if slices.Contains(PaymentTypes, il.Type) {
				metadata.paymentReceived = true
			} else if il.Type == "CREDIT WRITE OFF" && !writeOffReversed {
				metadata.contextType = "Write-off"
			}
		}
		if metadata.contextType == "" && il.Type == "CREDIT WRITE OFF" && !writeOffReversed {
			metadata.contextType = "Write-off pending"
		}
	}
}

func (ib *invoiceBuilder) addSupervisionLevels(supervisionLevels []store.GetSupervisionLevelsRow) {
	for _, sl := range supervisionLevels {
		ib.invoices[ib.invoicePositionByID[sl.InvoiceID.Int32]].invoice.SupervisionLevels = append(
			ib.invoices[ib.invoicePositionByID[sl.InvoiceID.Int32]].invoice.SupervisionLevels,
			shared.SupervisionLevel{
				Level:  sl.Supervisionlevel,
				Amount: int(sl.Amount),
				From:   shared.Date{Time: sl.Fromdate.Time},
				To:     shared.Date{Time: sl.Todate.Time},
			})
	}
}

func (s *Service) GetInvoices(ctx context.Context, clientId int32) (shared.Invoices, error) {
	invoices, err := s.store.GetInvoices(ctx, clientId)
	if err != nil {
		return nil, err
	}

	builder := newInvoiceBuilder(invoices)

	ledgerAllocations, err := s.store.GetLedgerAllocations(ctx, builder.GetIDs())

	if err != nil {
		s.Logger(ctx).Error("Get ledger allocations in get invoices has an issue " + err.Error())
		return shared.Invoices{}, err
	}

	builder.addLedgerAllocations(ledgerAllocations)
	supervisionLevels, err := s.store.GetSupervisionLevels(ctx, builder.GetIDs())

	if err != nil {
		s.Logger(ctx).Error("Get supervision levels in get invoices has an issue " + err.Error())
		return shared.Invoices{}, err
	}

	builder.addSupervisionLevels(supervisionLevels)
	return builder.Build(), nil
}
