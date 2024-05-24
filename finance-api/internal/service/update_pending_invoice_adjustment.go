package service

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"strings"
)

func (s *Service) UpdatePendingInvoiceAdjustment(ctx context.Context, ledgerId int, status string) error {
	logger := telemetry.LoggerFromContext(ctx)

	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(ctx); !errors.Is(err, sql.ErrTxDone) {
			logger.Error("Error rolling back update pending invoice adjustment transaction:", err)
		}
	}()

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
