package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"log/slog"
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

func (s *Service) UpdateClient(ctx context.Context, id int32, courtRef string) error {
	var cr pgtype.Text
	_ = cr.Scan(courtRef)

	return s.store.UpdateClient(ctx, store.UpdateClientParams{
		CourtRef: cr,
		ClientID: id,
	})
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
