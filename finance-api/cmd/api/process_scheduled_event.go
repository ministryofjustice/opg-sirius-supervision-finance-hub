package api

import (
	"context"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

const (
	scheduledEventRefundExpiry          = "refund-expiry"
	scheduledEventDirectDebitCollection = "direct-debit-collection"
)

func (s *Server) processScheduledEvent(ctx context.Context, event shared.ScheduledEvent) error {
	s.Logger(ctx).Info("processing scheduled event type " + event.Trigger)
	switch event.Trigger {
	case scheduledEventRefundExpiry:
		return s.service.ExpireRefunds(ctx)
	case scheduledEventDirectDebitCollection:
		collectionDate := time.Now().UTC().Truncate(24 * time.Hour)
		if detail, ok := event.Override.(shared.ScheduledDirectDebitCollectionOverride); ok {
			collectionDate = detail.Date.Time.UTC().Truncate(24 * time.Hour)
		}
		return s.service.AddCollectedPayments(ctx, collectionDate)
	default:
		return fmt.Errorf("invalid scheduled event trigger: %s", event.Trigger)
	}
}
