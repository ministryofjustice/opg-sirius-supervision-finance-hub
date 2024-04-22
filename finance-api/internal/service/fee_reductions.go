package service

import (
	"context"
	"github.com/opg-sirius-finance-hub/shared"
	"time"
)

func (s *Service) GetFeeReductions(id int) (*shared.FeeReductions, error) {
	ctx := context.Background()

	feeReductionsRawData, err := s.Store.GetFeeReductions(ctx, int32(id))
	if err != nil {
		return nil, err
	}

	var feeReductions shared.FeeReductions

	for _, fee := range feeReductionsRawData {
		startDate := shared.Date{Time: fee.Startdate.Time}
		endDate := shared.Date{Time: fee.Enddate.Time}
		var feeReduction = shared.FeeReduction{
			Id:              int(fee.ID),
			FinanceClientId: int(fee.FinanceClientID.Int32),
			Type:            fee.Type,
			StartDate:       startDate,
			EndDate:         endDate,
			DateReceived:    shared.Date{Time: fee.Datereceived.Time},
			Status:          calculateStatus(startDate, endDate, fee.Deleted),
			Notes:           fee.Notes,
		}
		feeReductions = append(feeReductions, feeReduction)
	}

	return &feeReductions, nil
}

func calculateStatus(startDate shared.Date, endDate shared.Date, deleted bool) string {
	now := shared.Date{Time: time.Now().Truncate(time.Hour * 24)}
	if deleted {
		return "Cancelled"
	} else if startDate.After(now) {
		return "Pending"
	} else if !now.Before(startDate) && !now.After(endDate) {
		return "Active"
	} else if endDate.Before(now) {
		return "Expired"
	}
	return ""
}
