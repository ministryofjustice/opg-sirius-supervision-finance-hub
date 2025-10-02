package allpay

import (
	"context"
)

type CancelPendingCollectionsRequest struct {
	ClientReference string
	Surname         string
}

func (c *Client) CancelPendingCollections(ctx context.Context, data *CancelPendingCollectionsRequest) error {

	return nil
}
