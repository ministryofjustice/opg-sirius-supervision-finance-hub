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
			Id:           int(fee.ID),
			Type:         fee.Type,
			StartDate:    startDate,
			EndDate:      endDate,
			DateReceived: shared.Date{Time: fee.Datereceived.Time},
			Status:       calculateStatus(startDate, endDate, fee.Deleted),
			Notes:        fee.Notes,
		}
		feeReduction.FeeReductionCancelAction = showFeeReductionCancelBtn(feeReduction.Status)
		feeReductions = append(feeReductions, feeReduction)
	}

	return &feeReductions, nil
}

func showFeeReductionCancelBtn(status any) bool {
	if status == shared.Pending || status == shared.Active {
		return true
	}
	return false
}

func calculateStatus(startDate shared.Date, endDate shared.Date, deleted bool) string {
	now := shared.Date{Time: time.Now().Truncate(time.Hour * 24)}
	if deleted {
		return shared.Cancelled
	} else if startDate.After(now) {
		return shared.Pending
	} else if !now.Before(startDate) && !now.After(endDate) {
		return shared.Active
	} else if endDate.Before(now) {
		return shared.Expired
	}
	return ""
}
