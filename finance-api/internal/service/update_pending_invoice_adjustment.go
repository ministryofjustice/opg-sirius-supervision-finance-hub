package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"strings"
)

func (s *Service) UpdatePendingInvoiceAdjustment(ctx context.Context, ledgerId int, status string) error {
	ctx, cancelTx := context.WithCancel(ctx)
	defer cancelTx()

	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return err
	}

	transaction := s.store.WithTx(tx)

	ledgerAdjustmentParams := store.UpdateLedgerAdjustmentParams{
		Status: strings.ToUpper(status),
		ID:     int32(ledgerId),
	}

	err = transaction.UpdateLedgerAdjustment(ctx, ledgerAdjustmentParams)
	if err != nil {
		return err
	}

	ledgerAllocationAdjustmentParams := store.UpdateLedgerAllocationAdjustmentParams(ledgerAdjustmentParams)

	err = transaction.UpdateLedgerAllocationAdjustment(ctx, ledgerAllocationAdjustmentParams)
	if err != nil {
		return err
	}

	ledger, err := transaction.GetLedger(ctx, int32(ledgerId))
	if err != nil {
		return err
	}

	if ledger.Type == "WRITE OFF REVERSAL" {
		// Reverse the write-off by adding the invoice amount back on
		_, err = transaction.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			LedgerID:  pgtype.Int4{int32(ledgerId), true},
			InvoiceID: ledger.InvoiceID,
			Amount:    -ledger.InvoiceAmount,
			Status:    "APPROVED",
			Notes:     ledger.Notes,
		})

		if err != nil {
			return err
		}

		// Move the credit from the ccb to the invoice
		_, err = transaction.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			LedgerID:  pgtype.Int4{int32(ledgerId), true},
			InvoiceID: ledger.InvoiceID,
			Amount:    ledger.Amount,
			Status:    "REAPPLIED",
			Notes:     ledger.Notes,
		})

		if err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
