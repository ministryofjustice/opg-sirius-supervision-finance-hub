package event

import (
	"context"
)

type FinanceAdminUploadFailed struct {
	EmailAddress string         `json:"emailAddress"`
	FailedLines  map[int]string `json:"failedLines"`
}

func (c *Client) FinanceAdminUploadFailed(ctx context.Context, event FinanceAdminUploadFailed) error {
	return c.send(ctx, "finance-admin-upload-failed", event)
}
