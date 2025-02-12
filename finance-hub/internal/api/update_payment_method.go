package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

func (c *ApiClient) SubmitPaymentMethod(ctx Context, clientId int, paymentMethod string) error {
	var body bytes.Buffer

	err := json.NewEncoder(&body).Encode(shared.UpdatePaymentMethod{
		PaymentMethod: shared.ParsePaymentMethod(paymentMethod),
	})

	if err != nil {
		return err
	}

	url := fmt.Sprintf("/clients/%d/payment-method", clientId)
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
	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	if resp.StatusCode != http.StatusNoContent {
		return newStatusError(resp)
	}

	return nil
}
