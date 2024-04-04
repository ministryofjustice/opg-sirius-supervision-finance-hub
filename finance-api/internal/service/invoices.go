package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"golang.org/x/sync/errgroup"
)

func (s *Service) GetInvoices(clientID int) (*shared.Invoices, error) {
	ctx := context.Background()
	group, _ := errgroup.WithContext(ctx)

	var invoicesData []store.GetInvoicesRow
	var ledgerAllocationsData []store.GetLedgerAllocationsRow
	var ledgerAllocations []shared.Ledger
	var invoices shared.Invoices
	var totalOfLedgerAllocationsAmount int

	group.Go(func() error {
		invoicesRawData, err := s.Store.GetInvoices(ctx, pgtype.Int4{Int32: int32(clientID), Valid: true})
		if err != nil {
			return err
		}
		invoicesData = invoicesRawData
		return nil
	})
	group.Go(func() error {
		ledgerAllocationsRawData, err := s.Store.GetLedgerAllocations(ctx, pgtype.Int4{Int32: int32(clientID), Valid: true})
		if err != nil {
			return err
		}
		ledgerAllocationsData = ledgerAllocationsRawData
		return nil
	})

	if err := group.Wait(); err != nil {
		return nil, err
	}

	for _, i := range ledgerAllocationsData {
		var ledgerAllocation = shared.Ledger{
			Amount:          int(i.Amount),
			ReceivedDate:    shared.Date{Time: i.Allocateddate.Time},
			TransactionType: "unknown",
			Status:          i.Status,
		}
		ledgerAllocations = append(ledgerAllocations, ledgerAllocation)
		totalOfLedgerAllocationsAmount = totalOfLedgerAllocationsAmount + int(i.Amount)
	}

	for _, i := range invoicesData {
		var invoice = shared.Invoice{
			Id:                 int(i.ID),
			Ref:                i.Reference,
			Status:             "",
			Amount:             int(i.Amount),
			RaisedDate:         shared.Date{Time: i.Raiseddate.Time},
			Received:           totalOfLedgerAllocationsAmount,
			OutstandingBalance: int(i.Amount) - totalOfLedgerAllocationsAmount,
			Ledgers:            ledgerAllocations,
			SupervisionLevels:  nil,
		}
		invoices = append(invoices, invoice)
	}

	return &invoices, nil
}
