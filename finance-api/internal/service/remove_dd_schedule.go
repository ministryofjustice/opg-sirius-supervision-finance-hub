package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) RemoveDirectDebitSchedule(ctx context.Context, data shared.RemoveSchedule) error {
	s.Logger(ctx).Info("removing schedule in Allpay", "courtRef", data.ClientReference)

	err := s.allpay.RemoveSchedule(ctx, &allpay.RemoveScheduleInput{
		ClosureDate: data.CollectionDate,
		Amount:      data.Amount,
		ClientDetails: allpay.ClientDetails{
			ClientReference: data.ClientReference,
			Surname:         data.Surname,
		},
	})
	if err != nil {
		return err
	}

	var (
		collectionDate pgtype.Date
		courtRef       pgtype.Text
	)
	_ = collectionDate.Scan(data.CollectionDate)
	_ = courtRef.Scan(data.ClientReference)
	_ = s.store.CancelPendingCollectionByCourtRefAndDate(ctx, store.CancelPendingCollectionByCourtRefAndDateParams{
		CourtRef:       courtRef,
		CollectionDate: collectionDate,
	})
	return nil
}
