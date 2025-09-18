package event

import (
	"context"
)

type DirectDebitScheduleFailed struct {
	ClientID int `json:"clientId"`
}

func (c *Client) DirectDebitScheduleFailed(ctx context.Context, event DirectDebitScheduleFailed) error {
	return c.send(ctx, "direct-debit-schedule-failed", event)
}
