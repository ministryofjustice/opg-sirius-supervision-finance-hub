package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"log"
)

func (s *Service) CreateLedgerEntry(clientId int, invoiceId int, ledgerEntry *shared.CreateLedgerEntryRequest) (string, error) {
	ctx := context.Background()

	balance, err := s.store.GetInvoiceBalance(context.Background(), int32(invoiceId))
	if err != nil {
		return "", err
	}

	err = s.validateAdjustmentAmount(ledgerEntry, balance)
	if err != nil {
		return "", err
	}

	params := store.CreateLedgerParams{
		Amount:    s.calculateAdjustmentAmount(ledgerEntry, balance),
		Notes:     pgtype.Text{String: ledgerEntry.AdjustmentNotes, Valid: true},
		Type:      ledgerEntry.AdjustmentType.Key(),
		InvoiceID: pgtype.Int4{Int32: int32(invoiceId), Valid: true},
		ClientID:  int32(clientId),
	}
	invoiceRef, err := s.store.CreateLedger(ctx, params)

	if err != nil {
		log.Println("Error creating ledger entry: ", err)
		return "", err
	}

	return invoiceRef, nil
}

func (s *Service) validateAdjustmentAmount(adjustment *shared.CreateLedgerEntryRequest, balance store.GetInvoiceBalanceRow) error {
	switch adjustment.AdjustmentType {
	case shared.AdjustmentTypeAddCredit:
		if int32(adjustment.Amount)-balance.Outstanding > balance.Initial {
			return shared.BadRequest{Field: "Amount", Reason: fmt.Sprintf("Amount entered must be equal to or less than £%d", (balance.Initial+balance.Outstanding)/100)}
		}
	case shared.AdjustmentTypeWriteOff:
		if int32(balance.Outstanding) < 1 {
			return shared.BadRequest{Field: "Amount", Reason: "No outstanding balance to write off"}
		}
	default:
		return shared.BadRequest{Field: "AdjustmentType", Reason: "Unimplemented adjustment type"}
	}
	return nil
}

func (s *Service) calculateAdjustmentAmount(adjustment *shared.CreateLedgerEntryRequest, balance store.GetInvoiceBalanceRow) int32 {
	switch adjustment.AdjustmentType {
	case shared.AdjustmentTypeWriteOff:
		return balance.Outstanding
	default:
		return int32(adjustment.Amount)
	}
}
