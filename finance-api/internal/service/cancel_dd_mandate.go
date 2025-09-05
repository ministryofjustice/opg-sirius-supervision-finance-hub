package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) CancelDirectDebitMandate(ctx context.Context, id int32, cancelMandate shared.CancelMandate) error {
	ctx, cancelTx := s.WithCancel(ctx)
	defer cancelTx()

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return err
	}

	// update payment method first, in case this fails
	err = tx.UpdatePaymentMethod(ctx, store.UpdatePaymentMethodParams{
		PaymentMethod: shared.PaymentMethodDemanded.Key(),
		ClientID:      id,
	})
	if err != nil {
		return err
	}

	// update allpay
	err = s.allpay.CancelMandate(ctx, &allpay.CancelMandateRequest{
		ClientReference: cancelMandate.CourtRef,
		Surname:         cancelMandate.Surname,
	})

	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Error cancelling mandate with allpay, rolling back payment method change for client : %d", id), slog.String("err", err.Error()))
		return err
	}

	err = s.dispatch.PaymentMethodChanged(ctx, event.PaymentMethod{
		ClientID:      int(id),
		PaymentMethod: shared.PaymentMethodDemanded,
	})
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
