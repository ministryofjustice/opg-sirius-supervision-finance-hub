package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) UpdatePendingInvoiceAdjustment(ctx context.Context, clientId int32, adjustmentId int32, status shared.AdjustmentStatus) error {
	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var updatedBy pgtype.Int4
	_ = store.ToInt4(&updatedBy, ctx.(auth.Context).User.ID)

	decisionParams := store.SetAdjustmentDecisionParams{
		ID:        adjustmentId,
		Status:    status.Key(),
		UpdatedBy: updatedBy,
	}

	adjustment, err := tx.SetAdjustmentDecision(ctx, decisionParams)
	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Set adjustment decision in updating invoice adjustment has an issue %s for client %d", err.Error(), clientId))

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

		ledgerID, err := tx.CreateLedgerForAdjustment(ctx, store.CreateLedgerForAdjustmentParams{
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
			s.Logger(ctx).Error(fmt.Sprintf("Error creating ledger for adjustment with id of %d for client %d", adjustmentId, clientId), slog.String("err", err.Error()))
			return err
		}

		for _, allocation := range allocations {
			allocation.LedgerID = ledgerID
			err = tx.CreateLedgerAllocation(ctx, allocation)
			if err != nil {
				s.Logger(ctx).Error(fmt.Sprintf("Error creating ledger allocation with id of %d for client %d", allocation.LedgerID, clientId), slog.String("err", err.Error()))
				return err
			}
		}
	}
	if status == shared.AdjustmentStatusApproved {
		err = s.ReapplyCredit(ctx, clientId, tx)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
