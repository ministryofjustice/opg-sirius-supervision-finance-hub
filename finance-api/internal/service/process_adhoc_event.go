package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) ProcessAdhocEvent(ctx context.Context, event shared.AdhocEvent) error {
	switch event.Task {
	case "ChangePendingCollectionDate":
		s.processAdhocChangePendingCollectionDate(ctx)
	case "SetDemanded":
		s.processAdhocSetDemanded(ctx)
	default:
		return fmt.Errorf("invalid adhoc process: %s", event.Task)
	}
	return nil
}

func (s *Service) processAdhocChangePendingCollectionDate(ctx context.Context) {
	// perform async so request context doesn't cancel before process is complete
	go func(logger *slog.Logger) {
		logger.Info("ChangePendingCollectionDate: started")

		//replaced context.Background with ctx to avoid cancellation, deadlines and logging context.
		funcCtx := telemetry.ContextWithLogger(ctx, logger)
		funcCtx = ctx.(auth.Context).WithContext(funcCtx)
		count, err := s.store.ChangePendingCollectionDate(funcCtx)
		if err != nil {
			logger.Error("ChangePendingCollectionDate: failed", "error", err)
			return
		}

		logger.Info(fmt.Sprintf("ChangePendingCollectionDate: %d pending collections updated", count))
	}(telemetry.LoggerFromContext(ctx))
}

func (s *Service) processAdhocSetDemanded(ctx context.Context) {
	// perform async so request context doesn't cancel before process is complete
	go func(logger *slog.Logger) {
		logger.Info("SetDemanded: started")

		clientIDs := []int{
			59290959,
			20051524,
			20044715,
			20083594,
			20098074,
			46667347,
			13475723,
			12884246,
			20090340,
			20043623,
			20051524,
			20044715,
			20083594,
			20098074,
			13811367,
			40131241,
			20114931,
			20090340,
			10478906,
			20104998,
		}

		//replaced context.Background with ctx to avoid cancellation, deadlines and logging context.
		funcCtx := telemetry.ContextWithLogger(ctx, logger)
		funcCtx = ctx.(auth.Context).WithContext(funcCtx)

		var count int
		for _, cid := range clientIDs {
			var (
				paymentMethod pgtype.Text
				id            pgtype.Int4
				createdBy     pgtype.Int4
			)
			_ = paymentMethod.Scan(shared.PaymentMethodDemanded.Key())
			_ = store.ToInt4(&id, cid)
			_ = store.ToInt4(&createdBy, 2180)

			// update payment method first, in case this fails
			err := s.store.SetPaymentMethod(funcCtx, store.SetPaymentMethodParams{
				PaymentMethod: paymentMethod,
				ClientID:      id,
				CreatedBy:     createdBy,
			})
			if err != nil {
				logger.Error("SetDemanded: failed", "error", err, "clientID", cid)
			}
			count++
		}

		logger.Info(fmt.Sprintf("SetDemanded: %d clients set to demanded", count))
	}(telemetry.LoggerFromContext(ctx))
}
