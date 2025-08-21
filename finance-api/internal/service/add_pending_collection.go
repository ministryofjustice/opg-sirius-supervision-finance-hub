package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) AddPendingCollection(ctx context.Context, clientId int32, data shared.PendingCollection) error {
	var collectionDate pgtype.Date
	
	_ = collectionDate.Scan(data.CollectionDate.Time)

	err := s.store.CreatePendingCollection(ctx, store.CreatePendingCollectionParams{
		ClientID:       clientId,
		CollectionDate: collectionDate,
		Amount:         int32(data.Amount),
		CreatedBy:      ctx.(auth.Context).User.ID,
	})

	return err
}
