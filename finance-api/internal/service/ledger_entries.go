package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"log"
)

func (s *Service) CreateLedgerEntry(clientId int, invoiceId int, ledgerEntry *shared.CreateLedgerEntryRequest) error {
	ctx := context.Background()

	err := s.validateAdjustmentAmount(invoiceId, ledgerEntry)
	if err != nil {
		return err
	}

	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err = tx.Rollback(ctx); err != nil {
			log.Println("Error rolling back transaction:", err)
		}
	}()

	q := s.store.WithTx(tx)

	params := store.CreateLedgerParams{
		Amount:    int32(ledgerEntry.Amount),
		Notes:     pgtype.Text{String: ledgerEntry.AdjustmentNotes, Valid: true},
		Type:      ledgerEntry.AdjustmentType.Key(),
		InvoiceID: pgtype.Int4{Int32: int32(invoiceId), Valid: true},
		ClientID:  int32(clientId),
	}
	err = q.CreateLedger(ctx, params)

	if err != nil {
		log.Println("Error creating ledger entry: ", err)
		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) validateAdjustmentAmount(invoiceId int, adjustment *shared.CreateLedgerEntryRequest) error {
	b, err := s.store.GetInvoiceBalance(context.Background(), int32(invoiceId))
	if err != nil {
		return err
	}
	switch adjustment.AdjustmentType {
	case shared.AdjustmentTypeAddCredit:
		if int32(adjustment.Amount)-b.Outstanding > b.Initial {
			return shared.BadRequest{Field: "Amount", Reason: fmt.Sprintf("Amount entered must be equal to or less than £%d", (b.Initial+b.Outstanding)/100)}
		}
	case shared.AdjustmentTypeWriteOff:
		if int32(b.Outstanding) < 1 {
			return shared.BadRequest{Field: "Amount", Reason: "No outstanding balance to write off"}
		}
	default:
		return shared.BadRequest{Field: "AdjustmentType", Reason: "Unimplemented adjustment type"}
	}
	return nil
}
