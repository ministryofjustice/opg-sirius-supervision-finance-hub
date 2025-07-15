package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

func (c *Client) GetBillingHistory(ctx context.Context, clientId int) ([]shared.BillingHistory, error) {
	var billingHistory []shared.BillingHistory

	url := fmt.Sprintf("/clients/%d/billing-history", clientId)
	req, err := c.newBackendRequest(ctx, http.MethodGet, url, nil)

	if err != nil {
		return billingHistory, err
	}

	resp, err := c.http.Do(req)

	if err != nil {
		return billingHistory, err
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode == http.StatusUnauthorized {
		return billingHistory, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return billingHistory, newStatusError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(&billingHistory)

	return billingHistory, err
}
