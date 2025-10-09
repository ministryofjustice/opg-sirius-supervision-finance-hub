package event

import (
	"context"
	"time"
)

type DirectDebitCollection struct {
	ClientID       int32     `json:"clientId"`
	Amount         int32     `json:"amount"`
	CollectionDate time.Time `json:"collectionDate"`
}

func (c *Client) DirectDebitCollection(ctx context.Context, event DirectDebitCollection) error {
	return c.send(ctx, "direct-debit-collection", event)
}
