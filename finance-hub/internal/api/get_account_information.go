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
	req, err := c.newBackendRequest(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		c.logger.Error("error building request", err)
		return v, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		c.logger.Error("error making request", err)
		return v, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return v, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("error from server on "+req.URL.Path, err)
		return v, newStatusError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(&v)
	return v, err
}
