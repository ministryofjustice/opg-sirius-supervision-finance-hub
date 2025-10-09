package service

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) PostReportActions(ctx context.Context, reportType shared.ReportRequest) {
	logger := s.Logger(ctx)
	switch reportType.ReportType {
	case shared.ReportsTypeDebt:
		switch *reportType.DebtType {
		case shared.DebtTypeApprovedRefunds:
			processed, err := s.store.MarkRefundsAsProcessed(ctx)
			if err != nil {
				logger.Error("post-report action to mark refunds as processed failed", "err", err)
			}
			logger.Info(fmt.Sprintf("post-report action - %d refunds marked as processed", len(processed)))
		default:
			return
		}
	default:
		return
	}
}
