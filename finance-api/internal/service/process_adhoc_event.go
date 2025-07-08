package service

import (
	"context"
	"fmt"
)

func (s *Service) ProcessAdhocEvent(ctx context.Context) error {
	logger := s.Logger(ctx)
	logger.Info("RebalanceCCB: process adhoc event")

	clientIDs, err := s.store.GetClientsWithCredit(ctx)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("RebalanceCCB: %d clients found", len(clientIDs)))

	for _, id := range clientIDs {
		err := s.ReapplyCredit(ctx, id, nil)
		if err != nil {
			logger.Error(fmt.Sprintf("RebalanceCCB: unable to reapply for client with id %d", id), "error", err)
			return err
		}
	}

	logger.Info("RebalanceCCB: run successfully")
	return nil
}
