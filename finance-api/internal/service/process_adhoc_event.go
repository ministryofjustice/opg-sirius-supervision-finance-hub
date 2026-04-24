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
		logger.Info("InvalidatePendingCollections: started")

		//replaced context.Background with ctx to avoid cancellation, deadlines and logging context.
		funcCtx := telemetry.ContextWithLogger(ctx, logger)
		funcCtx = ctx.(auth.Context).WithContext(funcCtx)
		count, err := s.store.PurgePendingCollections(funcCtx)
		if err != nil {
			logger.Error("InvalidatePendingCollections: failed", "error", err)
			return
		}

		logger.Info(fmt.Sprintf("InvalidatePendingCollections: %d pending collections deleted", count))
	}(telemetry.LoggerFromContext(ctx))

	return nil
}
