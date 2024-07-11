package api

import (
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

func (c *ApiClient) GetInvoiceAdjustments(ctx Context, clientId int) (shared.InvoiceAdjustments, error) {
	var invoiceAdjustments shared.InvoiceAdjustments

	url := fmt.Sprintf("/clients/%d/invoice-adjustments", clientId)

	req, err := c.newBackendRequest(ctx, http.MethodGet, url, nil)

	if err != nil {
		return invoiceAdjustments, err
	}

	resp, err := c.http.Do(req)

	if err != nil {
		return invoiceAdjustments, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return invoiceAdjustments, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return invoiceAdjustments, newStatusError(resp)
	}

	if err = json.NewDecoder(resp.Body).Decode(&invoiceAdjustments); err != nil {
		return invoiceAdjustments, err
	}

	return invoiceAdjustments, err
}
