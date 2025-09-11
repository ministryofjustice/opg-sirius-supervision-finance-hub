package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) UpdateRefundDecision(ctx context.Context, clientId int32, refundId int32, status shared.RefundStatus) error {
	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var (
		clientID  pgtype.Int4
		refundID  pgtype.Int4
		updatedBy pgtype.Int4
		decision  pgtype.Text
	)
	_ = store.ToInt4(&updatedBy, ctx.(auth.Context).User.ID)
	_ = store.ToInt4(&clientID, clientId)
	_ = store.ToInt4(&refundID, refundId)
	_ = decision.Scan(status.Key())

	switch decision.String {
	case shared.RefundStatusCancelled.Key():
		err = tx.CancelRefund(ctx, store.CancelRefundParams{
			CancelledBy: updatedBy,
			RefundID:    refundID,
			ClientID:    clientID,
		})
	case shared.RefundStatusApproved.Key(),
		shared.RefundStatusRejected.Key():
		decisionParams := store.SetRefundDecisionParams{
			Decision:   decision,
			DecisionBy: updatedBy,
			ClientID:   clientID,
			RefundID:   refundID,
		}
		err = tx.SetRefundDecision(ctx, decisionParams)
	default:
		err = errors.New("unknown decision type: " + decision.String)
	}

	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Set refund decision for client %d has error %s", clientId, err.Error()))
		return err
	}

	err = s.manageBankDetails(ctx, tx, refundId, status)
	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Removing bank details for refund %d has error %s", refundId, err.Error()))
		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) manageBankDetails(ctx context.Context, tx *store.Tx, refundId int32, status shared.RefundStatus) error {
	switch status {
	case shared.RefundStatusRejected,
		shared.RefundStatusCancelled,
		shared.RefundStatusFulfilled:
		return tx.RemoveBankDetails(ctx, refundId)
	default:
		return nil
	}
}
