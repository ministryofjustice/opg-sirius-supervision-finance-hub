package service

import (
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"log/slog"
)

func (s *Service) AddRefund(ctx context.Context, clientId int32, refund shared.AddRefund) error {
	refundableAmount, err := s.store.GetRefundAmount(ctx, clientId)
	if err != nil {
		return err
	}

	if refundableAmount == 0 {
		s.Logger(ctx).Error("Error creating refund with no funds to refund")
		return apierror.BadRequest{Reason: "NoCreditToRefund"}
	}

	_, err = s.store.CreateRefund(ctx, store.CreateRefundParams{
		ClientID:      clientId,
		Amount:        refundableAmount,
		Notes:         refund.RefundNotes,
		CreatedBy:     ctx.(auth.Context).User.ID,
		AccountName:   refund.AccountName,
		AccountNumber: refund.AccountNumber,
		SortCode:      refund.SortCode,
	})
	if err != nil {
		s.Logger(ctx).Error("Error creating refund", slog.String("err", err.Error()))
		return err
	}

	return s.dispatch.RefundAdded(ctx, event.RefundAdded{ClientID: clientId})
}
