package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) CancelFeeReduction(ctx context.Context, id int, cancelledFeeReduction shared.CancelFeeReduction) error {
	_, err := s.store.CancelFeeReduction(ctx, store.CancelFeeReductionParams{
		ID:                 int32(id),
		CancelledBy:        pgtype.Int4{Int32: int32(ctx.(auth.Context).User.ID), Valid: true},
		CancellationReason: pgtype.Text{String: cancelledFeeReduction.CancellationReason, Valid: true},
	})

	if err != nil {
		return err
	}

	return nil
}
