package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
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

func (s *Service) UpdateClient(ctx context.Context, id int, caseRecNumber string) error {
	err := s.store.UpdateClient(ctx, store.UpdateClientParams{
		Caserecnumber: pgtype.Text{String: caseRecNumber, Valid: true},
		ClientID:      int32(id),
	})

	if err != nil {
		return err
	}

	return nil
}
