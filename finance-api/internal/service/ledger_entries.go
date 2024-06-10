package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"log"
)

func (s *Service) CreateLedgerEntry(clientId int, invoiceId int, ledgerEntry *shared.CreateLedgerEntryRequest) (*shared.InvoiceReference, error) {
	ctx := context.Background()

	balance, err := s.store.GetInvoiceBalance(context.Background(), int32(invoiceId))
	if err != nil {
		return nil, err
	}

	err = s.validateAdjustmentAmount(ledgerEntry, balance)
	if err != nil {
		return nil, err
	}

	params := store.CreateLedgerParams{
		Amount:    s.calculateAdjustmentAmount(ledgerEntry, balance),
		Notes:     pgtype.Text{String: ledgerEntry.AdjustmentNotes, Valid: true},
		Type:      ledgerEntry.AdjustmentType.Key(),
		InvoiceID: pgtype.Int4{Int32: int32(invoiceId), Valid: true},
		ClientID:  int32(clientId),
	}
	invoiceReference, err := s.store.CreateLedger(ctx, params)

	if err != nil {
		log.Println("Error creating ledger entry: ", err)
		return nil, err
	}

	return &shared.InvoiceReference{InvoiceRef: invoiceReference}, nil
}

func (s *Service) validateAdjustmentAmount(adjustment *shared.CreateLedgerEntryRequest, balance store.GetInvoiceBalanceRow) error {
	switch adjustment.AdjustmentType {
	case shared.AdjustmentTypeAddCredit:
		if int32(adjustment.Amount)-balance.Outstanding > balance.Initial {
			return shared.BadRequest{Field: "Amount", Reason: fmt.Sprintf("Amount entered must be equal to or less than £%s", shared.IntToDecimalString(int(balance.Initial+balance.Outstanding)))}
		}
	case shared.AdjustmentTypeAddDebit:
		var maxBalance int32
		if balance.Feetype == "AD" {
			maxBalance = 10000
		} else {
			maxBalance = 32000
		}
		if int32(adjustment.Amount)+balance.Outstanding > maxBalance {
			return shared.BadRequest{Field: "Amount", Reason: fmt.Sprintf("Amount entered must be equal to or less than £%s", shared.IntToDecimalString(int(maxBalance-balance.Outstanding)))}
		}
	case shared.AdjustmentTypeWriteOff:
		if balance.Outstanding < 1 {
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
	case shared.AdjustmentTypeAddDebit:
		return -int32(adjustment.Amount)
	default:
		return int32(adjustment.Amount)
	}
}
