package service

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"log/slog"
)

func (s *Service) GetRefunds(ctx context.Context, clientId int32) (shared.Refunds, error) {
	var refunds shared.Refunds

	ai, err := s.store.GetAccountInformation(ctx, clientId)

	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Error in getting account information for client %d", clientId), slog.String("err", err.Error()))
		return refunds, err
	}

	data, err := s.store.GetRefunds(ctx, clientId)

	if err != nil {
		s.Logger(ctx).Error("Error get refunds", slog.String("err", err.Error()))
		return refunds, err
	}

	refunds.CreditBalance = int(ai.Credit)

	for _, refund := range data {
		var r = shared.Refund{
			ID:            int(refund.ID),
			RaisedDate:    shared.Date{Time: refund.RaisedDate.Time},
			FulfilledDate: shared.TransformNillablePgDate(refund.FulfilledDate),
			Amount:        int(refund.Amount),
			Status:        shared.ParseRefundStatus(refund.Status),
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

		refunds.Refunds = append(refunds.Refunds, r)
	}

	return refunds, nil
}
