package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"golang.org/x/exp/maps"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"slices"
)

var (
	FeeReductionTypes = []string{"CREDIT EXEMPTION", "CREDIT HARDSHIP", "CREDIT REMISSION"}
	AllocatedStatuses = []string{"ALLOCATED", "APPROVED", "CONFIRMED"} // TODO: PFS-107 & PFS-110 to investigate
)

type invoiceMetadata struct {
	invoice     *shared.Invoice
	total       int
	contextType string
	creditAdded bool
}

type invoiceBuilder struct {
	invoices map[int32]*invoiceMetadata
}

func newInvoiceBuilder(invoices []store.GetInvoicesRow) *invoiceBuilder {
	ib := invoiceBuilder{make(map[int32]*invoiceMetadata)}
	for _, inv := range invoices {
		ib.invoices[inv.ID] = &invoiceMetadata{
			invoice: &shared.Invoice{
				Id:                 int(inv.ID),
				Ref:                inv.Reference,
				Amount:             int(inv.Amount),
				RaisedDate:         shared.Date{Time: inv.Raiseddate.Time},
				Status:             "Unpaid",
				OutstandingBalance: int(inv.Amount),
			},
		}
	}
	return &ib
}

func (ib *invoiceBuilder) IDs() []int32 {
	return maps.Keys(ib.invoices)
}

func (ib *invoiceBuilder) Invoices() *shared.Invoices {
	var invoices shared.Invoices
	for _, inv := range ib.invoices {
		invoices = append(invoices, *inv.invoice)
	}
	return &invoices
}

func (ib *invoiceBuilder) addLedgerAllocations(ilas []store.GetLedgerAllocationsRow) {
	ledgersByInvoice := map[int32][]shared.Ledger{}

	for _, il := range ilas {
		ledgersByInvoice[il.InvoiceID.Int32] = append(
			ledgersByInvoice[il.InvoiceID.Int32],
			shared.Ledger{
				Amount:          int(il.Amount),
				ReceivedDate:    calculateReceivedDate(il.Bankdate, il.Datetime),
				TransactionType: il.Type,
				Status:          il.Status,
			})

		metadata := ib.invoices[il.InvoiceID.Int32]
		if slices.Contains(AllocatedStatuses, il.Status) {
			metadata.total += int(il.Amount)
			if slices.Contains(FeeReductionTypes, il.Type) && metadata.contextType == "" {
				metadata.contextType = cases.Title(language.English).String(il.Type)
			}
			if il.Type == "CREDIT MEMO" {
				metadata.creditAdded = true
			} else if il.Type == "CREDIT WRITE OFF" {
				metadata.contextType = "Write-off"
			}
		}
		if metadata.contextType == "" && il.Type == "CREDIT WRITE OFF" {
			metadata.contextType = "Write-off pending"
		}
	}

	for id, ledgers := range ledgersByInvoice {
		metadata := ib.invoices[id]
		invoice := metadata.invoice
		invoice.Ledgers = ledgers
		invoice.Received = metadata.total
		invoice.OutstandingBalance -= metadata.total

		var status string
		if invoice.OutstandingBalance > 0 {
			status = "Unpaid"
		} else if invoice.OutstandingBalance < 0 {
			status = "Overpaid"
		} else if invoice.OutstandingBalance == 0 {
			if metadata.contextType != "" || metadata.creditAdded {
				status = "Closed"
			} else {
				status = "Paid"
			}
		}
		invoice.Status = status

		if metadata.contextType != "" {
			invoice.Status = fmt.Sprintf("%s - %s", invoice.Status, metadata.contextType)
		}
	}
}

func (ib *invoiceBuilder) addSupervisionLevels(supervisionLevels []store.GetSupervisionLevelsRow) {
	supervisionLevelsByInvoice := map[int32][]shared.SupervisionLevel{}
	for _, sl := range supervisionLevels {
		supervisionLevelsByInvoice[sl.InvoiceID.Int32] = append(
			supervisionLevelsByInvoice[sl.InvoiceID.Int32],
			shared.SupervisionLevel{
				Level:  sl.Supervisionlevel,
				Amount: int(sl.Amount),
				From:   shared.Date{Time: sl.Fromdate.Time},
				To:     shared.Date{Time: sl.Todate.Time},
			})
	}

	for id, sls := range supervisionLevelsByInvoice {
		ib.invoices[id].invoice.SupervisionLevels = sls
	}
}

func (s *Service) GetInvoices(clientID int) (*shared.Invoices, error) {
	ctx := context.Background()

	invoices, err := s.store.GetInvoices(ctx, int32(clientID))
	if err != nil {
		return nil, err
	}

	builder := newInvoiceBuilder(invoices)

	ledgerAllocations, err := s.store.GetLedgerAllocations(ctx, builder.IDs())
	if err != nil {
		return nil, err
	}

	builder.addLedgerAllocations(ledgerAllocations)

	supervisionLevels, err := s.store.GetSupervisionLevels(ctx, builder.IDs())
	if err != nil {
		return nil, err
	}

	builder.addSupervisionLevels(supervisionLevels)

	return builder.Invoices(), nil
}

func calculateReceivedDate(bankDate pgtype.Date, datetime pgtype.Timestamp) shared.Date {
	if !bankDate.Time.IsZero() {
		return shared.Date{Time: bankDate.Time}
	}

	return shared.Date{Time: datetime.Time}
}
