package event

import (
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type PaymentMethod struct {
	ClientID      int            `json:"clientId"`
	PaymentMethod shared.RefData `json:"paymentMethod"`
}

func (c *Client) PaymentMethod(ctx context.Context, event PaymentMethod) error {
	return c.send(ctx, "payment-method", event)
}
