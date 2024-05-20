package service

import (
	"context"
	"log"
)

func (s *Service) UpdatePendingInvoiceAdjustment(ledgerId int) error {
	ctx := context.Background()

	tx, err := s.TX.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			log.Println("Error rolling back update pending invoice adjustment transaction:", err)
		}
	}()

	transaction := s.Store.WithTx(tx)
	ledgerIdTransformed := int32(ledgerId)

	_, err = transaction.UpdateLedgerAdjustment(ctx, ledgerIdTransformed)
	if err != nil {
		return err
	}

	_, err = transaction.UpdateLedgerAllocationAdjustment(ctx, ledgerIdTransformed)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
