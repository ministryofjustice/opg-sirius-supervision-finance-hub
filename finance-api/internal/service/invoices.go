package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/shared"
)

var totalOfLedgerAllocationsAmount int

func (s *Service) GetInvoices(clientID int) (*shared.Invoices, error) {
	ctx := context.Background()

	var invoices shared.Invoices

	invoicesRawData, err := s.Store.GetInvoices(ctx, pgtype.Int4{Int32: int32(clientID), Valid: true})
	if err != nil {
		return nil, err
	}

	for _, i := range invoicesRawData {
		ledgerAllocations, totalOfLedgerAllocationsAmount, err := getLedgerAllocations(s, ctx, int(i.ID))
		if err != nil {
			return nil, err
		}
		var invoice = shared.Invoice{
			Id:                 int(i.ID),
			Ref:                i.Reference,
			Status:             "",
			Amount:             int(i.Amount),
			RaisedDate:         shared.Date{Time: i.Raiseddate.Time},
			Received:           totalOfLedgerAllocationsAmount,
			OutstandingBalance: int(i.Amount) - totalOfLedgerAllocationsAmount,
			Ledgers:            ledgerAllocations,
			SupervisionLevels:  getSupervisionLevels(s, ctx, int(i.ID)),
		}

		invoices = append(invoices, invoice)
	}

	return &invoices, nil
}

func getSupervisionLevels(s *Service, ctx context.Context, invoiceID int) []shared.SupervisionLevel {
	var supervisionLevels []shared.SupervisionLevel
	supervisionLevelsRawData, err := s.Store.GetSupervisionLevels(ctx, pgtype.Int4{Int32: int32(invoiceID), Valid: true})
	if err != nil {
		return nil
	}

	for _, i := range supervisionLevelsRawData {
		var supervisionLevel = shared.SupervisionLevel{
			Level:  i.Supervisionlevel,
			Amount: int(i.Amount),
			From:   shared.Date{Time: i.Fromdate.Time},
			To:     shared.Date{Time: i.Todate.Time},
		}
		supervisionLevels = append(supervisionLevels, supervisionLevel)
	}
	return supervisionLevels
}

func getLedgerAllocations(s *Service, ctx context.Context, invoiceID int) ([]shared.Ledger, int, error) {
	var ledgerAllocations []shared.Ledger
	totalOfLedgerAllocationsAmount = 0
	ledgerAllocationsRawData, err := s.Store.GetLedgerAllocations(ctx, pgtype.Int4{Int32: int32(invoiceID), Valid: true})
	if err != nil {
		return nil, 0, err
	}

	for _, i := range ledgerAllocationsRawData {
		var ledgerAllocation = shared.Ledger{
			Amount:          int(i.Amount),
			ReceivedDate:    calculateReceivedDate(i.Bankdate, i.Datetime),
			TransactionType: i.Type,
			Status:          "Applied",
		}
		ledgerAllocations = append(ledgerAllocations, ledgerAllocation)
		totalOfLedgerAllocationsAmount = totalOfLedgerAllocationsAmount + int(i.Amount)
	}
	return ledgerAllocations, totalOfLedgerAllocationsAmount, nil
}

func calculateReceivedDate(bankDate pgtype.Date, datetime pgtype.Timestamp) shared.Date {
	if !bankDate.Time.IsZero() {
		return shared.Date{Time: bankDate.Time}
	}

	return shared.Date{Time: datetime.Time}
}
