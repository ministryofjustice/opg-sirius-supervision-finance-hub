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
		return v, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return v, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return v, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return v, newStatusError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(&v)
	return v, err
}
