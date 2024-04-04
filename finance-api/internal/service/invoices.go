package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/shared"
)

func (s *Service) GetInvoices(clientID int) (*shared.Invoices, error) {
	ctx := context.Background()

	var ledgerAllocations []shared.Ledger
	var invoices shared.Invoices
	var totalOfLedgerAllocationsAmount int

	var id pgtype.Int4
	err := id.Scan(int64(clientID))
	if err != nil {
		return nil, err
	}

	invoicesRawData, err := s.Store.GetInvoices(ctx, id)
	if err != nil {
		return nil, err
	}

	ledgerAllocationsRawData, err := s.Store.GetLedgerAllocations(ctx, id)
	if err != nil {
		return nil, err
	}

	for _, i := range ledgerAllocationsRawData {
		var ledgerAllocation = shared.Ledger{
			Amount:          int(i.Amount),
			ReceivedDate:    shared.Date{Time: i.Allocateddate.Time},
			TransactionType: "unknown",
			Status:          i.Status,
		}
		ledgerAllocations = append(ledgerAllocations, ledgerAllocation)
		totalOfLedgerAllocationsAmount = totalOfLedgerAllocationsAmount + int(i.Amount)
	}

	for _, i := range invoicesRawData {
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
