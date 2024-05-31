package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"golang.org/x/exp/maps"
)

type invoiceBuilder struct {
	invoices map[int32]*shared.Invoice
}

func newInvoiceBuilder(invoices []store.GetInvoicesRow) *invoiceBuilder {
	ib := invoiceBuilder{make(map[int32]*shared.Invoice)}
	for _, inv := range invoices {
		ib.invoices[inv.ID] = &shared.Invoice{
			Id:         int(inv.ID),
			Ref:        inv.Reference,
			Amount:     int(inv.Amount),
			RaisedDate: shared.Date{Time: inv.Raiseddate.Time},
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
		invoices = append(invoices, *inv)
	}
	return &invoices
}

func (ib *invoiceBuilder) addLedgerAllocations(ilas []store.GetLedgerAllocationsRow) {
	ledgersByInvoice := map[int32][]shared.Ledger{}
	totalByInvoice := map[int32]int{}

	for _, il := range ilas {
		ledgersByInvoice[il.InvoiceID.Int32] = append(
			ledgersByInvoice[il.InvoiceID.Int32],
			shared.Ledger{
				Amount:          int(il.Amount),
				ReceivedDate:    calculateReceivedDate(il.Bankdate, il.Datetime),
				TransactionType: il.Type,
				Status:          il.Status,
			})
		if il.Status == "APPROVED" {
			totalByInvoice[il.InvoiceID.Int32] += int(il.Amount)
		}
	}

	for id, ledgers := range ledgersByInvoice {
		invoice := ib.invoices[id]
		invoice.Ledgers = ledgers
		invoice.Received = totalByInvoice[id]
		invoice.OutstandingBalance = invoice.Amount - totalByInvoice[id]
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
		ib.invoices[id].SupervisionLevels = sls
	}
}

func (ib *invoiceBuilder) calculateStatus() {
	// do me
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

	builder.calculateStatus()

	return builder.Invoices(), nil
}

func calculateReceivedDate(bankDate pgtype.Date, datetime pgtype.Timestamp) shared.Date {
	if !bankDate.Time.IsZero() {
		return shared.Date{Time: bankDate.Time}
	}

	return shared.Date{Time: datetime.Time}
}
