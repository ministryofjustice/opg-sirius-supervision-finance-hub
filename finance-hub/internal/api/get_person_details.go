package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"

	"net/http"
)

func (c *Client) GetPersonDetails(ctx context.Context, ClientId int) (shared.Person, error) {
	var v shared.Person

	requestURL := fmt.Sprintf("/clients/%d", ClientId)
	req, err := c.newSiriusRequest(ctx, http.MethodGet, requestURL, nil)
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
