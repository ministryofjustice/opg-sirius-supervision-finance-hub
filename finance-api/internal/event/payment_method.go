package event

import (
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type PaymentMethod struct {
	ClientID      int                  `json:"clientId"`
	PaymentMethod shared.PaymentMethod `json:"paymentMethod"`
}

func (c *Client) PaymentMethodChanged(ctx context.Context, event PaymentMethod) error {
	return c.send(ctx, "payment-method-changed", event)
}
