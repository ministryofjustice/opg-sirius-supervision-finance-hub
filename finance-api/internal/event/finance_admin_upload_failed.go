package event

import (
	"context"
)

type FinanceAdminUploadProcessed struct {
	EmailAddress string         `json:"emailAddress"`
	FailedLines  map[int]string `json:"failedLines"`
	Error        string         `json:"error"`
}

func (c *Client) FinanceAdminUploadProcessed(ctx context.Context, event FinanceAdminUploadProcessed) error {
	return c.send(ctx, "finance-admin-upload-processed", event)
}
