package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"time"
)

func (s *Service) AddFeeReduction(body shared.AddFeeReduction) error {
	ctx := context.Background()

	dateString := body.DateReceive
	date, _ := time.Parse("2006-01-02", dateString)

	queryArgs := store.AddFeeReductionParams{
		FinanceClientID: pgtype.Int4{Int32: int32(body.FinanceClientId)},
		Type:            body.FeeType,
		Evidencetype:    pgtype.Text{},
		Startdate:       pgtype.Date{Time: date},
		Enddate:         pgtype.Date{Time: date},
		Notes:           body.FeeReductionNotes,
		Deleted:         false,
		Datereceived:    pgtype.Date{Time: date},
	}

	_, err := s.Store.AddFeeReduction(ctx, queryArgs)
	if err != nil {
		return err
	}

	return nil
}
