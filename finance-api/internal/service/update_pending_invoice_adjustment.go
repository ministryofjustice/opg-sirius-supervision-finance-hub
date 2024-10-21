package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
)

func (s *Service) UpdatePendingInvoiceAdjustment(ctx context.Context, clientId int, adjustmentId int, status shared.AdjustmentStatus) error {
	ctx, cancelTx := context.WithCancel(ctx)
	defer cancelTx()

	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return err
	}

	transaction := s.store.WithTx(tx)

	decisionParams := store.SetAdjustmentDecisionParams{
		ID:        int32(adjustmentId),
		Status:    status.Key(),
		UpdatedBy: pgtype.Int4{Int32: int32(1), Valid: true},
	}

	adjustment, err := transaction.SetAdjustmentDecision(ctx, decisionParams)
	if err != nil {
		return err
	}

	if status == shared.AdjustmentStatusApproved {
		ledger, allocations := generateLedgerEntries(addLedgerVars{
			amount:             adjustment.Amount,
			transactionType:    shared.ParseAdjustmentType(adjustment.AdjustmentType),
			clientId:           int32(clientId),
			invoiceId:          adjustment.InvoiceID,
			outstandingBalance: adjustment.Outstanding,
		})
		ledgerId, err := transaction.CreateLedger(ctx, ledger)
		if err != nil {
			return err
		}

		for _, allocation := range allocations {
			allocation.LedgerID = pgtype.Int4{Int32: ledgerId, Valid: true}
			err = transaction.CreateLedgerAllocation(ctx, allocation)
			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	if status == shared.AdjustmentStatusApproved {
		return s.ReapplyCredit(ctx, int32(clientId))
	}
	return nil
}
