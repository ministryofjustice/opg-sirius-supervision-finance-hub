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

	invoicesRawData, err := s.Store.GetInvoices(ctx, pgtype.Int4{Int32: int32(clientID), Valid: true})
	if err != nil {
		return nil, err
	}

	ledgerAllocationsRawData, err := s.Store.GetLedgerAllocations(ctx, pgtype.Int4{Int32: int32(clientID), Valid: true})
	if err != nil {
		return nil, err
	}

	for _, i := range ledgerAllocationsRawData {
		var ledgerAllocation = shared.Ledger{
			Amount:          int(i.Amount),
			ReceivedDate:    CheckForAllocation(i.Allocateddate, i.Datetime),
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

func CheckForAllocation(allocatedDate pgtype.Date, datetime pgtype.Timestamp) shared.Date {
	if !allocatedDate.Time.IsZero() {
		return shared.Date{Time: allocatedDate.Time}
	}

	return shared.Date{Time: datetime.Time}
}
