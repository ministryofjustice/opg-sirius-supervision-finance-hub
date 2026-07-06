package event

import "context"

type RefundReset struct {
	ClientID int32 `json:"clientId"`
}

func (c *Client) RefundReset(ctx context.Context, event RefundReset) error {
	return c.send(ctx, "refund-reset", event)
}
