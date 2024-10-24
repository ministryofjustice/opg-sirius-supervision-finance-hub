package event

import (
	"context"
)

type FinanceAdminUploadProcessed struct {
	EmailAddress string         `json:"emailAddress"`
	FailedLines  map[int]string `json:"failedLines"`
	Error        string         `json:"error"`
	ReportType   string         `json:"reportType"`
}

func (c *Client) FinanceAdminUploadProcessed(ctx context.Context, event FinanceAdminUploadProcessed) error {
	return c.send(ctx, "finance-admin-upload-processed", event)
}
