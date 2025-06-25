package api

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

const (
	scheduledEventRefundExpiry = "refund-expiry"
)

func (s *Server) processScheduledEvent(ctx context.Context, event shared.ScheduledEvent) error {
	switch event.Trigger {
	case scheduledEventRefundExpiry:
		return s.service.ExpireRefunds(ctx)
	default:
		return fmt.Errorf("invalid scheduled event trigger: %s", event.Trigger)
	}
}
