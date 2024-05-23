package api

import (
	"fmt"
	"net/http"
)

func (c *ApiClient) UpdatePendingInvoiceAdjustment(ctx Context, clientId int, ledgerId int) error {

	url := fmt.Sprintf("/clients/%d/invoice-adjustments/%d", clientId, ledgerId)
	req, err := c.newBackendRequest(ctx, http.MethodPut, url, nil)

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
