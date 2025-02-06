package event

import (
	"context"
)

type PaymentMethod struct {
	ClientID      int    `json:"clientId"`
	PaymentMethod string `json:"paymentMethod"`
}

func (c *Client) PaymentMethod(ctx context.Context, event PaymentMethod) error {
	return c.send(ctx, "payment-method", event)
}
