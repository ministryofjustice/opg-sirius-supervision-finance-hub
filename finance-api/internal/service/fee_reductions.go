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

	for _, fee := range feeReductionsRawData {
		startDate := shared.Date{Time: fee.Startdate.Time}
		endDate := shared.Date{Time: fee.Enddate.Time}
		var feeReduction = shared.FeeReduction{
			Id:           int(fee.ID),
			Type:         fee.Type,
			StartDate:    startDate,
			EndDate:      endDate,
			DateReceived: shared.Date{Time: fee.Datereceived.Time},
			Status:       calculateStatus(startDate, endDate, fee.Deleted),
			Notes:        fee.Notes,
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
