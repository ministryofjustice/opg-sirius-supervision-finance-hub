package sirius

import (
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/internal/model"
	"net/http"
)

func (c *ApiClient) GetInvoices(ctx Context, clientId int) (model.Invoices, error) {
	var invoices model.Invoices

	url := fmt.Sprintf("/api/v1/clients/%d/invoices", clientId)

	req, err := c.newRequest(ctx, http.MethodGet, url, nil)

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
