package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
)

func (s *Service) ExpireRefunds(ctx context.Context) error {
	s.Logger(ctx).Info("starting refund expiry job")

	var user pgtype.Int4
	_ = store.ToInt4(&user, ctx.(auth.Context).User.ID)

	pendingCount, err := s.store.ExpirePendingRefunds(ctx, user)
	if err != nil {
		s.Logger(ctx).Error("pending refund expiry failed", "error", err)
		return err
	}
	s.Logger(ctx).Info(fmt.Sprintf("%d expired pending refunds set to rejected", pendingCount))

	approvedCount, err := s.store.ExpireApprovedRefunds(ctx, user)
	if err != nil {
		s.Logger(ctx).Error("approved refund expiry failed", "error", err)
		return err
	}
	s.Logger(ctx).Info(fmt.Sprintf("%d expired approved refunds set to cancelled", approvedCount))

	processingCount, err := s.store.ExpireProcessingRefunds(ctx, user)
	if err != nil {
		s.Logger(ctx).Error("processing refund expiry failed", "error", err)
		return err
	}
	s.Logger(ctx).Info(fmt.Sprintf("%d expired processing refunds set to cancelled", processingCount))

	return nil
}
