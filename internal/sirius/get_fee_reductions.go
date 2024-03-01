package sirius

import (
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/internal/model"

	"net/http"
)

func (c *ApiClient) GetFeeReductions(ctx Context, clientId int) (model.FeeReductions, error) {
	var v model.FeeReductions

	requestURL := fmt.Sprintf("/api/v1/clients/%d/fee-reductions", clientId)
	req, err := c.newRequest(ctx, http.MethodGet, requestURL, nil)
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