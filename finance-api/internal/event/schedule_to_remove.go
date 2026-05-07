package event

import (
	"context"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type ScheduleToRemove struct {
	CourtRef string      `json:"courtRef"`
	Surname  string      `json:"surname"`
	Amount   int32       `json:"amount"`
	Date     shared.Date `json:"date"`
}

func (c *Client) ScheduleToRemove(ctx context.Context, event ScheduleToRemove) error {
	return c.send(ctx, "schedule-to-remove", event)
}
