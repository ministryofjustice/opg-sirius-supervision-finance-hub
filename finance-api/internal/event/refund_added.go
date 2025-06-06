package event

import "context"

type RefundAdded struct {
	ClientID int32 `json:"clientId"`
}

func (c *Client) RefundAdded(ctx context.Context, event RefundAdded) error {
	return c.send(ctx, "refund-added", event)
}
