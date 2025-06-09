package service

import (
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"log/slog"
)

func (s *Service) GetRefunds(ctx context.Context, clientId int32) (shared.Refunds, error) {
	var refunds shared.Refunds

	data, err := s.store.GetRefunds(ctx, clientId)

	if err != nil {
		s.Logger(ctx).Error("Error get refunds", slog.String("err", err.Error()))
		return nil, err
	}

	for _, refund := range data {
		var r = shared.Refund{
			ID:            int(refund.ID),
			RaisedDate:    shared.Date{Time: refund.RaisedDate.Time},
			FulfilledDate: shared.TransformNillablePgDate(refund.FulfilledDate),
			Amount:        int(refund.Amount),
			Status:        refund.Status,
			Notes:         refund.Notes,
			BankDetails: shared.Nillable[shared.BankDetails]{
				Value: shared.BankDetails{
					Name:     refund.AccountName,
					Account:  refund.AccountCode,
					SortCode: refund.SortCode,
				},
				Valid: refund.AccountName != "" && refund.AccountCode != "" && refund.SortCode != "",
			},
			CreatedBy: int(refund.CreatedBy),
		}

		refunds = append(refunds, r)
	}

	return refunds, nil
}
