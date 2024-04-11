package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/shared"
	"time"
)

func (s *Service) GetFeeReductions(id int) (*shared.FeeReductions, error) {
	ctx := context.Background()

	feeReductionsRawData, err := s.Store.GetFeeReductions(ctx, pgtype.Int4{Int32: int32(id), Valid: true})
	if err != nil {
		return nil, err
	}

	var feeReductions shared.FeeReductions

	for _, f := range feeReductionsRawData {
		startDate := shared.Date{Time: f.Startdate.Time}
		endDate := shared.Date{Time: f.Enddate.Time}
		var feeReduction = shared.FeeReduction{
			Id:           int(f.ID),
			Type:         f.Discounttype,
			StartDate:    startDate,
			EndDate:      endDate,
			DateReceived: shared.Date{Time: f.Datereceived.Time},
			Status:       calculateStatus(startDate, endDate, f.Deleted),
			Notes:        f.Notes,
		}
		feeReductions = append(feeReductions, feeReduction)
	}

	return &feeReductions, nil
}

func calculateStatus(startDate shared.Date, endDate shared.Date, deleted bool) string {
	now := shared.Date{Time: time.Now()}
	if deleted {
		return "Cancelled"
	} else if startDate.Before(now) && endDate.After(now) {
		return "Active"
	} else if endDate.Before(now) {
		return "Expired"
	}
	return ""
}
