package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) CancelFeeReduction(ctx context.Context, id int32, cancelledFeeReduction shared.CancelFeeReduction) error {
	var (
		cancelledBy        pgtype.Int4
		cancellationReason pgtype.Text
	)

	_ = store.ToInt4(&cancelledBy, ctx.(auth.Context).User.ID)
	_ = cancellationReason.Scan(cancelledFeeReduction.CancellationReason)

	_, err := s.store.CancelFeeReduction(ctx, store.CancelFeeReductionParams{
		ID:                 id,
		CancelledBy:        cancelledBy,
		CancellationReason: cancellationReason,
	})

	if err != nil {
		return err
	}

	return nil
}
