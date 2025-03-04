package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) UpdatePendingInvoiceAdjustment(ctx context.Context, clientId int32, adjustmentId int32, status shared.AdjustmentStatus) error {
	ctx, cancelTx := s.WithCancel(ctx)
	defer cancelTx()

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return err
	}

	var updatedBy pgtype.Int4
	_ = store.ToInt4(&updatedBy, ctx.(auth.Context).User.ID)

	decisionParams := store.SetAdjustmentDecisionParams{
		ID:        adjustmentId,
		Status:    status.Key(),
		UpdatedBy: updatedBy,
	}

	adjustment, err := tx.SetAdjustmentDecision(ctx, decisionParams)
	if err != nil {
		return err
	}

	if status == shared.AdjustmentStatusApproved {
		ledger, allocations := generateLedgerEntries(ctx, addLedgerVars{
			amount:             adjustment.Amount,
			transactionType:    shared.ParseAdjustmentType(adjustment.AdjustmentType),
			clientId:           clientId,
			invoiceId:          adjustment.InvoiceID,
			outstandingBalance: adjustment.Outstanding,
		})

		id, err := tx.CreateLedgerForAdjustment(ctx, store.CreateLedgerForAdjustmentParams{
			ClientID:       ledger.ClientID,
			Amount:         ledger.Amount,
			Notes:          ledger.Notes,
			Type:           ledger.Type,
			Status:         ledger.Status,
			FeeReductionID: ledger.FeeReductionID,
			CreatedBy:      ledger.CreatedBy,
			ID:             adjustmentId,
		})
		if err != nil {
			return err
		}

		var ledgerID pgtype.Int4
		_ = store.ToInt4(&ledgerID, id)

		for _, allocation := range allocations {
			allocation.LedgerID = ledgerID
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
		return s.ReapplyCredit(ctx, clientId)
	}
	return nil
}
