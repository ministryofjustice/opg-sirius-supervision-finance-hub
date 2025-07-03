package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"

	"net/http"
)

func (c *Client) GetPermittedAdjustments(ctx context.Context, clientId int, invoiceId int) ([]shared.AdjustmentType, error) {
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

	defer unchecked(resp.Body.Close)

	if resp.StatusCode == http.StatusUnauthorized {
		return types, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return types, newStatusError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(&types)
	return types, err
}
