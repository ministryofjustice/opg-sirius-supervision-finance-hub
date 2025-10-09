package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

func (c *Client) GetFeeReductions(ctx context.Context, clientId int) (shared.FeeReductions, error) {
	var v shared.FeeReductions

	url := fmt.Sprintf("/clients/%d/fee-reductions", clientId)
	req, err := c.newBackendRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return v, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return v, err
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode == http.StatusUnauthorized {
		return v, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return v, newStatusError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(&v)
	return v, err
}
