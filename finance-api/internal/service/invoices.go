package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/shared"
	"regexp"
	"strconv"
)

func (s *Service) GetInvoices(id int) (*shared.Invoices, error) {
	ctx := context.Background()

	fc, err := s.Store.GetInvoices(ctx, pgtype.Int4{Int32: int32(id), Valid: true})
	if err != nil {
		return nil, err
	}

	ledgerAllocationsData, err := s.Store.GetLedgerAllocations(ctx, pgtype.Int4{Int32: int32(id), Valid: true})
	if err != nil {
		return nil, err
	}

	var ledgerAllocations []shared.Ledger
	var invoices shared.Invoices
	var totalOfLedgerAllocationsAmount int

	for _, i := range ledgerAllocationsData {
		var ledgerAllocation = shared.Ledger{
			Amount:          IntToDecimalString(int(i.Amount)),
			ReceivedDate:    shared.Date{Time: i.Allocateddate.Time},
			TransactionType: "unknown",
			Status:          i.Status,
		}
		ledgerAllocations = append(ledgerAllocations, ledgerAllocation)
		totalOfLedgerAllocationsAmount = totalOfLedgerAllocationsAmount + int(i.Amount)
	}

	for _, i := range fc {
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

func IntToDecimalString(i int) string {
	s := strconv.FormatFloat(float64(i)/100, 'f', -1, 32)
	const singleDecimal = "\\.\\d$"
	m, _ := regexp.Match(singleDecimal, []byte(s))
	if m {
		s = s + "0"
	}
	return s
}
