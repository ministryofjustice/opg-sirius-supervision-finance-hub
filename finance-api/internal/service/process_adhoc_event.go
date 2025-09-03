package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
)

func (s *Service) ProcessAdhocEvent(ctx context.Context) error {
	// perform async so request context doesn't cancel before process is complete
	go func(logger *slog.Logger) {
		logger.Info("RebalanceCCB: process adhoc event")
		funcCtx := telemetry.ContextWithLogger(context.Background(), logger)
		funcCtx = ctx.(auth.Context).WithContext(funcCtx)
		clientIDs, err := s.store.GetClientsWithCredit(funcCtx)
		if err != nil {
			logger.Error("RebalanceCCB: unable to fetch clients", "error", err)
			return
		}

		logger.Info(fmt.Sprintf("RebalanceCCB: %d clients found", len(clientIDs)))

		for _, id := range clientIDs {
			err := s.ReapplyCredit(funcCtx, id, nil)
			if err != nil {
				logger.Error(fmt.Sprintf("RebalanceCCB: unable to reapply for client with id %d", id), "error", err)
			}
		}

		logger.Info("RebalanceCCB: run successfully")
	}(telemetry.LoggerFromContext(ctx))

	return nil
}
