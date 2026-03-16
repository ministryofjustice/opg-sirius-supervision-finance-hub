package event

import (
	"context"
)

type PendingInvoiceAdjustment struct {
	ClientID         int    `json:"clientId"`
	AdjustmentType   string `json:"adjustmentType"`
	InvoiceReference string `json:"InvoiceReference"`
}

func (c *Client) PendingInvoiceAdjustment(ctx context.Context, event PendingInvoiceAdjustment) error {
	return c.send(ctx, "pending-invoice-adjustment", event)
}
