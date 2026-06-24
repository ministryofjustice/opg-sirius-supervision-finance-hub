package api

import (
	"context"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Server) processAdhocEvent(ctx context.Context, event shared.AdhocEvent) error {
	err := s.service.ProcessAdhocEvent(ctx, event)
	if err != nil {
		return err
	}
	return nil
}
