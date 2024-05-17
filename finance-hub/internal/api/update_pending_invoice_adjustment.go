package api

import (
	"fmt"
	"net/http"
)

func (c *ApiClient) UpdatePendingInvoiceAdjustment(ctx Context, clientId int, invoiceId int, adjustmentType string, notes string, amount string) error {

	url := fmt.Sprintf("/clients/%d/invoice-adjustments/%d", clientId, invoiceId)
	req, err := c.newBackendRequest(ctx, http.MethodPost, url, nil)

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

	//if resp.StatusCode != http.StatusCreated {
	//	var v shared.ValidationError
	//
	//	if err := json.NewDecoder(resp.Body).Decode(&v); err == nil && len(v.Errors) > 0 {
	//		return shared.ValidationError{Errors: v.Errors}
	//	}
	//
	//	return newStatusError(resp)
	//}

	return nil
}
