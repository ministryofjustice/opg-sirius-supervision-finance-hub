package event

import (
	"context"
)

type DirectDebitMandateReview struct {
	ClientID int32 `json:"clientId"`
}

func (c *Client) DirectDebitMandateReview(ctx context.Context, event DirectDebitMandateReview) error {
	return c.send(ctx, "direct-debit-mandate-review", event)
}
