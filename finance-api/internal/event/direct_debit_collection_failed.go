package event

import (
	"context"
)

type DirectDebitCollectionFailed struct {
	ClientID int `json:"clientId"`
}

func (c *Client) DirectDebitCollectionFailed(ctx context.Context, event DirectDebitCollectionFailed) error {
	return c.send(ctx, "direct-debit-collection-failed", event)
}
