package service

import (
	"context"
	"strconv"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) QueueScheduleRemovals(ctx context.Context, schedules [][]string, scheduleDate shared.Date) {
	for i, schedule := range schedules {
		if i == 0 { // skip header
			continue
		}
		amount, _ := strconv.Atoi(schedule[2])
		err := s.dispatch.ScheduleToRemove(ctx, event.ScheduleToRemove{
			CourtRef: schedule[0],
			Surname:  schedule[1],
			Amount:   amount,
			Date:     scheduleDate,
		})
		if err != nil {
			s.Logger(ctx).Error("failed to queue schedule removal", "err", err, "category", "allpay")
		}
	}
}
