package event

import "context"

type CreditOnAccount struct {
	ClientID        int `json:"clientId"`
	CreditRemaining int `json:"creditRemaining"`
}

func (c *Client) CreditOnAccount(ctx context.Context, event CreditOnAccount) error {
	return c.send(ctx, "credit-on-account", event)
}
