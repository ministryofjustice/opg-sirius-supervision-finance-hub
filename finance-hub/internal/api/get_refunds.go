package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

func (c *Client) GetRefunds(ctx context.Context, clientId int) (refunds shared.Refunds, err error) {
	url := fmt.Sprintf("/clients/%d/refunds", clientId)

	req, err := c.newBackendRequest(ctx, http.MethodGet, url, nil)

	if err != nil {
		return refunds, err
	}

	resp, err := c.http.Do(req)

	if err != nil {
		return refunds, err
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode == http.StatusUnauthorized {
		return refunds, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return refunds, newStatusError(resp)
	}

	if err = json.NewDecoder(resp.Body).Decode(&refunds); err != nil {
		return refunds, err
	}

	return refunds, err
}
