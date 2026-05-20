package api

import (
	"context"
	"fmt"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Server) processScheduledEvent(ctx context.Context, event shared.ScheduledEvent) error {
	s.Logger(ctx).Info("processing scheduled event type " + event.Trigger)
	switch event.Trigger {
	case shared.ScheduledEventRefundExpiry:
		return s.service.ExpireRefunds(ctx)
	default:
		return fmt.Errorf("invalid scheduled event trigger: %s", event.Trigger)
	}
}
