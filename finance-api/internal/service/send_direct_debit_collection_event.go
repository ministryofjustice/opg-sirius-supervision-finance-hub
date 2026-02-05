package service

import (
	"context"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
)

func (s *Service) SendDirectDebitCollectionEvent(ctx context.Context, clientID int32, pendingCollection PendingCollection) error {
	return s.dispatch.DirectDebitCollection(ctx, event.DirectDebitCollection{
		ClientID:       clientID,
		Amount:         pendingCollection.Amount,
		CollectionDate: pendingCollection.CollectionDate,
	})
}
