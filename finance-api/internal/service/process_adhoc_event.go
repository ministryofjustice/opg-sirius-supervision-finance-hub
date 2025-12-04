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
		logger.Info("UpdateRefundLedgerAmounts: started")
		funcCtx := telemetry.ContextWithLogger(context.Background(), logger)
		funcCtx = ctx.(auth.Context).WithContext(funcCtx)
		count, err := s.store.UpdateRefundLedgerAmounts(funcCtx)
		if err != nil {
			logger.Error("UpdateRefundLedgerAmounts: failed", "error", err)
			return
		}

		logger.Info(fmt.Sprintf("UpdateRefundLedgerAmounts: %d ledgers updated", count))
	}(telemetry.LoggerFromContext(ctx))

	return nil
}
