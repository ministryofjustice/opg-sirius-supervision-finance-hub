package event

import (
	"context"
)

type DebtChaseUploaded struct {
	UserID   int32  `json:"userId"`
	Filename string `json:"filename"`
}

func (c *Client) DebtChaseUploaded(ctx context.Context, event DebtChaseUploaded) error {
	return c.send(ctx, "debt-chase-uploaded", event)
}
