package service

import (
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) AddRefund(ctx context.Context, clientId int32, data shared.BankDetails) error {
	s.store.CreateRefund
}
