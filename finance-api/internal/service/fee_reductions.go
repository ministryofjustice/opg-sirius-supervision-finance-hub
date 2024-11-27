package service

import (
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"time"
)

func (s *Service) GetFeeReductions(ctx context.Context, id int) (*shared.FeeReductions, error) {
	feeReductionsRawData, err := s.store.GetFeeReductions(ctx, int32(id))
	if err != nil {
		return nil, err
	}

	var feeReductions shared.FeeReductions

	for _, fee := range feeReductionsRawData {
		startDate := shared.Date{Time: fee.Startdate.Time}
		endDate := shared.Date{Time: fee.Enddate.Time}
		var feeReduction = shared.FeeReduction{
			Id:           int(fee.ID),
			Type:         shared.ParseFeeReductionType(fee.Type),
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
	now := shared.Date{Time: time.Now().Truncate(time.Hour * 24)}
	if deleted {
		return shared.StatusCancelled
	} else if startDate.After(now) {
		return shared.StatusPending
	} else if !now.Before(startDate) && !now.After(endDate) {
		return shared.StatusActive
	} else if endDate.Before(now) {
		return shared.StatusExpired
	}
	return ""
}
