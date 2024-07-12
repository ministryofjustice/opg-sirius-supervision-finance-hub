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
		return types, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return types, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return types, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return types, newStatusError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(&types)
	return types, err
}
