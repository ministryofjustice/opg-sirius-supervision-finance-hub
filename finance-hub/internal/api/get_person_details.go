package api

import (
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/shared"

	"net/http"
)

func (c *ApiClient) GetPersonDetails(ctx Context, ClientId int) (shared.Person, error) {
	var v shared.Person

	requestURL := fmt.Sprintf("/api/v1/clients/%d", ClientId)
	req, err := c.newSiriusRequest(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		c.logErrorRequest(req, err)
		return v, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		c.logger.Request(req, err)
		return v, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		c.logger.Request(req, err)
		return v, shared.ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Request(req, err)
		return v, newStatusError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(&v)
	return v, err
}
