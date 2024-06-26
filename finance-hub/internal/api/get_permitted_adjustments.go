package api

import (
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/shared"

	"net/http"
)

func (c *ApiClient) GetPermittedAdjustments(ctx Context, clientId int, invoiceId int) ([]shared.AdjustmentType, error) {
	var types []shared.AdjustmentType

	requestURL := fmt.Sprintf("/clients/%d/invoices/%d/permitted-adjustments", clientId, invoiceId)
	req, err := c.newBackendRequest(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		c.logErrorRequest(req, err)
		return types, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		c.logger.Request(req, err)
		return types, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		c.logger.Request(req, err)
		return types, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Request(req, err)
		return types, newStatusError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(&types)
	return types, err
}
