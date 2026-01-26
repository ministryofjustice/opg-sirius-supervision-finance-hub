package service

import (
	"context"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
)

func (s *Service) CheckPaymentMethod(ctx context.Context, clientID int32) error {
	method, err := s.store.GetPaymentMethod(ctx, clientID)
	if err != nil {
		return err
	}
	if method == "DIRECT DEBIT" {
		return s.dispatch.DirectDebitMandateReview(ctx, event.DirectDebitMandateReview{ClientID: clientID})
	}
	return nil
}
