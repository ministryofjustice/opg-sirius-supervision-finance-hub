package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) GetAccountInformation(ctx context.Context, id int32) (*shared.AccountInformation, error) {
	fc, err := s.store.GetAccountInformation(ctx, id)

	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Error in getting account information for client %d", id), slog.String("err", err.Error()))
		return nil, err
	}

	return &shared.AccountInformation{
		CreditBalance:      int(fc.Credit),
		OutstandingBalance: int(fc.Outstanding),
		PaymentMethod:      fc.PaymentMethod,
	}, nil
}


func (s *Service) UpdatePaymentMethod(ctx context.Context, id int32, paymentMethod shared.PaymentMethod) error {
	err := s.store.UpdatePaymentMethod(ctx, store.UpdatePaymentMethodParams{
		PaymentMethod: paymentMethod.Key(),
		ClientID:      id,
	})

	if err == nil {
		return s.dispatch.PaymentMethodChanged(ctx, event.PaymentMethod{
			ClientID:      int(id),
			PaymentMethod: paymentMethod,
		})
	}
	return err
}
