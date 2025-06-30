package api

import (
	"context"
	"encoding/json"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

func (c *Client) GetUserSession(ctx context.Context) (*shared.User, error) {
	req, err := c.newSessionRequest(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return nil, newStatusError(resp)
	}

	var v shared.User
	err = json.NewDecoder(resp.Body).Decode(&v)
	return &v, err
}
