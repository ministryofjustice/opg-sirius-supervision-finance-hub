package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/shared"
)

func (s *Service) GetInvoices(clientID int) (*shared.Invoices, error) {
	ctx := context.Background()
	var invoices shared.Invoices

	invoicesRawData, err := s.Store.GetInvoices(ctx, pgtype.Int4{Int32: int32(clientID), Valid: true})
	if err != nil {
		return nil, err
	}

	for _, inv := range invoicesRawData {
		ledgerAllocations, totalOfLedgerAllocationsAmount, err := getLedgerAllocations(s, ctx, int(inv.ID))
		if err != nil {
			return nil, err
		}

		supervisionLevels, err := getSupervisionLevels(s, ctx, int(inv.ID))
		if err != nil {
			return nil, err
		}

		var invoice = shared.Invoice{
			Id:                 int(inv.ID),
			Ref:                inv.Reference,
			Status:             "",
			Amount:             int(inv.Amount),
			RaisedDate:         shared.Date{Time: inv.Raiseddate.Time},
			Received:           totalOfLedgerAllocationsAmount,
			OutstandingBalance: int(inv.Amount) - totalOfLedgerAllocationsAmount,
			Ledgers:            ledgerAllocations,
			SupervisionLevels:  supervisionLevels,
		}

		invoices = append(invoices, invoice)
	}

	return &invoices, nil
}

func getSupervisionLevels(s *Service, ctx context.Context, invoiceID int) ([]shared.SupervisionLevel, error) {
	var supervisionLevels []shared.SupervisionLevel
	supervisionLevelsRawData, err := s.Store.GetSupervisionLevels(ctx, pgtype.Int4{Int32: int32(invoiceID), Valid: true})
	if err != nil {
		return nil, err
	}

	for _, supervision := range supervisionLevelsRawData {
		var supervisionLevel = shared.SupervisionLevel{
			Level:  supervision.Supervisionlevel,
			Amount: int(supervision.Amount),
			From:   shared.Date{Time: supervision.Fromdate.Time},
			To:     shared.Date{Time: supervision.Todate.Time},
		}
		supervisionLevels = append(supervisionLevels, supervisionLevel)
	}
	return supervisionLevels, nil
}

func getLedgerAllocations(s *Service, ctx context.Context, invoiceID int) ([]shared.Ledger, int, error) {
	var ledgerAllocations []shared.Ledger
	totalOfLedgerAllocationsAmount := 0
	ledgerAllocationsRawData, err := s.Store.GetLedgerAllocations(ctx, pgtype.Int4{Int32: int32(invoiceID), Valid: true})
	if err != nil {
		return nil, 0, err
	}

	for _, ledger := range ledgerAllocationsRawData {
		var ledgerAllocation = shared.Ledger{
			Amount:          int(ledger.Amount),
			ReceivedDate:    calculateReceivedDate(ledger.Bankdate, ledger.Datetime),
			TransactionType: ledger.Type,
			Status:          "Applied",
		}
		ledgerAllocations = append(ledgerAllocations, ledgerAllocation)
		totalOfLedgerAllocationsAmount = totalOfLedgerAllocationsAmount + int(ledger.Amount)
	}
	return ledgerAllocations, totalOfLedgerAllocationsAmount, nil
}

func calculateReceivedDate(bankDate pgtype.Date, datetime pgtype.Timestamp) shared.Date {
	if !bankDate.Time.IsZero() {
		return shared.Date{Time: bankDate.Time}
	}

	return shared.Date{Time: datetime.Time}
}
