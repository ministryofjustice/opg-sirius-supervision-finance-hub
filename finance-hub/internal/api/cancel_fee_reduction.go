package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/apierror"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

func (c *ApiClient) CancelFeeReduction(ctx Context, clientId int, feeReductionId int, cancellationReason string) error {
	var body bytes.Buffer

	err := json.NewEncoder(&body).Encode(shared.CancelFeeReduction{
		CancellationReason: cancellationReason,
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("/clients/%d/fee-reductions/%d/cancel", clientId, feeReductionId)
	req, err := c.newBackendRequest(ctx, http.MethodPut, url, &body)

	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
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
