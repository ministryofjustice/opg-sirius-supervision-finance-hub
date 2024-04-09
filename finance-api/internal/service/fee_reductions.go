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

	for _, i := range feeReductionsRawData {
		var feeReduction = shared.FeeReduction{
			Id:           int(i.ID),
			Type:         i.Discounttype,
			StartDate:    shared.Date{Time: i.Startdate.Time},
			EndDate:      shared.Date{Time: i.Enddate.Time},
			DateReceived: shared.Date{Time: i.Datereceived.Time},
			Status:       i.Status,
			Notes:        i.Notes,
		}
		feeReductions = append(feeReductions, feeReduction)
	}

	return &feeReductions, nil
}
