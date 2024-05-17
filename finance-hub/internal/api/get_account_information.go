package api

import (
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/shared"

	"net/http"
)

func (c *ApiClient) GetAccountInformation(ctx Context, ClientId int) (shared.AccountInformation, error) {
	var v shared.AccountInformation

	requestURL := fmt.Sprintf("/clients/%d", ClientId)
	req, err := c.newRequest(ctx, http.MethodGet, requestURL, nil, "financeApi")
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
		return v, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Request(req, err)
		return v, newStatusError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(&v)
	return v, err
}
