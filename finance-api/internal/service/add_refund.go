package service

import (
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"log/slog"
	"strconv"
)

func (s *Service) AddRefund(ctx context.Context, clientId int32, refund shared.AddRefund) error {
	amount, err := s.store.GetRefundAmount(ctx, clientId)
	if err != nil {
		return err
	}

	if amount == 0 {
		s.Logger(ctx).Error("Error creating refund with no funds to refund")
		return apierror.BadRequest{Reason: "NoCreditToRefund"}
	}

	r, err := s.store.CreateRefund(ctx, store.CreateRefundParams{
		ClientID:  clientId,
		Amount:    amount,
		Notes:     refund.Notes,
		CreatedBy: ctx.(auth.Context).User.ID,
		Name:      refund.AccountName,
		Account:   refund.Account,
		SortCode:  refund.SortCode,
	})
	if err != nil {
		s.Logger(ctx).Error("Error creating refund", slog.String("err", err.Error()))
		return err
	}
	s.Logger(ctx).Debug(strconv.Itoa(int(r)))
	refunds, err := s.store.GetRefunds(ctx, clientId)
	if len(refunds) > 0 {

	}

	return s.dispatch.RefundAdded(ctx, event.RefundAdded{ClientID: clientId})
}
