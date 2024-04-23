package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
)

func (s *Service) AddFeeReduction(body shared.AddFeeReduction) error {
	ctx := context.Background()

	queryArgs := store.AddFeeReductionParams{
		ClientID:     int32(body.ClientId),
		Type:         body.FeeType,
		Evidencetype: pgtype.Text{},
		Startdate:    pgtype.Date{Time: body.DateReceive.Time, Valid: true},
		Enddate:      pgtype.Date{Time: body.DateReceive.Time, Valid: true},
		Notes:        body.FeeReductionNotes,
		Deleted:      false,
		Datereceived: pgtype.Date{Time: body.DateReceive.Time, Valid: true},
	}

	_, err := s.Store.AddFeeReduction(ctx, queryArgs)
	if err != nil {
		return err
	}

	return nil
}
