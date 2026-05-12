package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) QueueScheduleRemovals(ctx context.Context, schedules [][]string, scheduleDate shared.Date) map[int]string {
	failedLines := make(map[int]string)

	for i, schedule := range schedules {
		if i == 0 { // skip header
			continue
		}

		var courtRef pgtype.Text

		_ = courtRef.Scan(schedule[0])

		_, err := s.store.GetClientByCourtRef(ctx, courtRef)
		if err != nil {
			failedLines[i] = validation.UploadErrorClientNotFound
			continue
		}

		amount, err := parseAmount(safeRead(schedule, 2))
		if err != nil {
			failedLines[i] = validation.UploadErrorAmountParse
			continue
		}

		err = s.dispatch.ScheduleToRemove(ctx, event.ScheduleToRemove{
			CourtRef: schedule[0],
			Surname:  schedule[1],
			Amount:   amount,
			Date:     scheduleDate,
		})
		if err != nil {
			s.Logger(ctx).Error("failed to queue schedule removal", "err", err, "category", "allpay")
		}
	}
	return failedLines
}
