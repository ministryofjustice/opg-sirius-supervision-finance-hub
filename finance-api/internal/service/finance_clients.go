package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) GetAccountInformation(ctx context.Context, id int) (*shared.AccountInformation, error) {
	fc, err := s.store.GetAccountInformation(ctx, int32(id))
	if err != nil {
		return nil, err
	}

	return &shared.AccountInformation{
		CreditBalance:      int(fc.Credit),
		OutstandingBalance: int(fc.Outstanding),
		PaymentMethod:      fc.PaymentMethod,
	}, nil
}

func (s *Service) UpdateClient(ctx context.Context, id int, courtRef string) error {
	return s.store.UpdateClient(ctx, store.UpdateClientParams{
		CourtRef: pgtype.Text{String: courtRef, Valid: true},
		ClientID: int32(id),
	})
}
