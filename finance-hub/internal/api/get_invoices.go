package api

import (
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
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
		return invoices, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return invoices, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return invoices, newStatusError(resp)
	}

	if err = json.NewDecoder(resp.Body).Decode(&invoices); err != nil {
		return invoices, err
	}

	return invoices, err
}
