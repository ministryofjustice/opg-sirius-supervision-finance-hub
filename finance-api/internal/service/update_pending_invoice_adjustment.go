package service

import (
	"context"
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

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
