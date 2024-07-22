package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
)

func (s *Service) CancelFeeReduction(ctx context.Context, id int, cancelledFeeReduction shared.CancelFeeReduction) error {
	_, err := s.store.CancelFeeReduction(ctx, store.CancelFeeReductionParams{
		ID: int32(id),
		//TODO make sure we have correct createdby ID in ticket PFS-88
		CancelledBy:        pgtype.Int4{Int32: 1, Valid: true},
		CancellationReason: pgtype.Text{String: cancelledFeeReduction.CancellationReason, Valid: true},
	})

	if err != nil {
		return err
	}

	return nil
}
