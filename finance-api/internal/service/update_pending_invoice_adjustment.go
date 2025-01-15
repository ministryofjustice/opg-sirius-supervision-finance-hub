package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) UpdatePendingInvoiceAdjustment(ctx context.Context, clientId int, adjustmentId int, status shared.AdjustmentStatus) error {
	ctx, cancelTx := context.WithCancel(ctx)
	defer cancelTx()

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return err
	}

	decisionParams := store.SetAdjustmentDecisionParams{
		ID:        int32(adjustmentId),
		Status:    status.Key(),
		UpdatedBy: pgtype.Int4{Int32: int32(1), Valid: true},
	}

	adjustment, err := tx.SetAdjustmentDecision(ctx, decisionParams)
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

		ledgerId, err := tx.CreateLedgerForAdjustment(ctx, store.CreateLedgerForAdjustmentParams{
			ClientID:       ledger.ClientID,
			Amount:         ledger.Amount,
			Notes:          ledger.Notes,
			Type:           ledger.Type,
			Status:         ledger.Status,
			FeeReductionID: ledger.FeeReductionID,
			CreatedBy:      ledger.CreatedBy,
			ID:             int32(adjustmentId),
		})
		if err != nil {
			return err
		}

		for _, allocation := range allocations {
			allocation.LedgerID = pgtype.Int4{Int32: ledgerId, Valid: true}
			err = tx.CreateLedgerAllocation(ctx, allocation)
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
