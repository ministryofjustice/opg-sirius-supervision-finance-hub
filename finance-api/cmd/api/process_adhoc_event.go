package api

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Server) processAdhocEvent(ctx context.Context, event shared.AdhocEvent) error {
	if event.Task != "RebalanceCCB" {
		return fmt.Errorf("invalid adhoc process: %s", event.Task)
	}

	err := s.service.ProcessAdhocEvent(ctx)
	if err != nil {
		return err
	}
	return nil
}
