package api

import (
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

func (c *ApiClient) GetBillingHistory(ctx Context, clientId int) ([]shared.BillingHistory, error) {
	var billingHistory []shared.BillingHistory

	url := fmt.Sprintf("/clients/%d/billing-history", clientId)

	req, err := c.newBackendRequest(ctx, http.MethodGet, url, nil)

	if err != nil {
		return billingHistory, err
	}

	resp, err := c.http.Do(req)

	if err != nil {
		c.logger.Request(req, err)
		return billingHistory, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		c.logger.Request(req, err)
		return billingHistory, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Request(req, err)
		return billingHistory, newStatusError(resp)
	}

	if err = json.NewDecoder(resp.Body).Decode(&billingHistory); err != nil {
		c.logger.Request(req, err)
		return billingHistory, err
	}

	return billingHistory, err
}
