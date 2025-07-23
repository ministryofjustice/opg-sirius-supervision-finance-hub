package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"net/http"
)

func (c *Client) CancelDirectDebit(ctx context.Context, clientId int) error {
	url := fmt.Sprintf("/clients/%d/direct-debits/cancel", clientId)
	req, err := c.newBackendRequest(ctx, http.MethodPost, url, nil)

	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusAccepted {
		return nil
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	if resp.StatusCode == http.StatusUnprocessableEntity {
		var v apierror.ValidationError
		if err := json.NewDecoder(resp.Body).Decode(&v); err == nil && len(v.Errors) > 0 {
			return apierror.ValidationError{Errors: v.Errors}
		}
	}

	return newStatusError(resp)
}
