package api

import (
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

func (c *ApiClient) GetInvoice(ctx Context, clientId int, invoiceId int) (shared.Invoice, error) {
	var invoice shared.Invoice

	url := fmt.Sprintf("/clients/%d/invoices/%d", clientId, invoiceId)

	req, err := c.newBackendRequest(ctx, http.MethodGet, url, nil)

	if err != nil {
		return invoice, err
	}

	resp, err := c.http.Do(req)

	if err != nil {
		c.logger.Request(req, err)
		return invoice, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		c.logger.Request(req, err)
		return invoice, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Request(req, err)
		return invoice, newStatusError(resp)
	}

	if err = json.NewDecoder(resp.Body).Decode(&invoice); err != nil {
		c.logger.Request(req, err)
		return invoice, err
	}

	return invoice, err
}
