package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *ApiClient) UpdatePendingInvoiceAdjustment(ctx Context, clientId int, ledgerId int, status string) error {
	var body bytes.Buffer

	err := json.NewEncoder(&body).Encode(struct {
		Status string `json:"status"`
	}{
		Status: status,
	})

	url := fmt.Sprintf("/clients/%d/invoice-adjustments/%d", clientId, ledgerId)
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
