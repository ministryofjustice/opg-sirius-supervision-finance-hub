package api

import (
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

func (c *ApiClient) GetInvoices(ctx Context, clientId int) (shared.Invoices, error) {
	var invoices shared.Invoices

	url := fmt.Sprintf("/clients/%d/invoices", clientId)

	req, err := c.newBackendRequest(ctx, http.MethodGet, url, nil)

	if err != nil {
		return invoices, err
	}

	resp, err := c.http.Do(req)

	if err != nil {
		c.logger.Request(req, err)
		return invoices, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		c.logger.Request(req, err)
		return invoices, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Request(req, err)
		return invoices, newStatusError(resp)
	}

	if err = json.NewDecoder(resp.Body).Decode(&invoices); err != nil {
		c.logger.Request(req, err)
		return invoices, err
	}

	return invoices, err
}
