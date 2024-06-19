package service

import (
	"context"
	"github.com/opg-sirius-finance-hub/shared"
)

func (s *Service) GetAccountInformation(id int) (*shared.AccountInformation, error) {
	ctx := context.Background()

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
