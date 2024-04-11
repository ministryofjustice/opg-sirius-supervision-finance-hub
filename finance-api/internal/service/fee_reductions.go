package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/shared"
)

func (s *Service) GetFeeReductions(id int) (*shared.FeeReductions, error) {
	ctx := context.Background()

	feeReductionsRawData, err := s.Store.GetFeeReductions(ctx, pgtype.Int4{Int32: int32(id), Valid: true})
	if err != nil {
		return nil, err
	}

	var feeReductions shared.FeeReductions

	for _, f := range feeReductionsRawData {
		var feeReduction = shared.FeeReduction{
			Id:           int(f.ID),
			Type:         f.Discounttype,
			StartDate:    shared.Date{Time: f.Startdate.Time},
			EndDate:      shared.Date{Time: f.Enddate.Time},
			DateReceived: shared.Date{Time: f.Datereceived.Time},
			Status:       f.Status,
			Notes:        f.Notes,
		}
		feeReductions = append(feeReductions, feeReduction)
	}

	return &feeReductions, nil
}
