package service

import (
	"context"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) UpdateClientMandateDetails(ctx context.Context, id int32, detail shared.ClientUpdatedEvent) error {
	client, err := s.store.GetClientById(ctx, id)
	if err != nil {
		return err
	}

	if client.PaymentMethod != shared.PaymentMethodDirectDebit.Key() {
		return nil
	}

	s.Logger(ctx).Info("updating client details in Allpay", "courtRef", client.CourtRef)

	input := &allpay.UpdateClientDetailsInput{
		ClientDetails: allpay.ClientDetails{
			ClientReference: client.CourtRef,
			Surname:         detail.Surname.Old,
		},
		NewSurname: detail.Surname.New,
		Address: allpay.Address{
			Line1:    client.Line1,
			Town:     client.Town,
			PostCode: client.Postcode,
		},
	}
	return s.allpay.UpdateClientDetails(ctx, input)
}
