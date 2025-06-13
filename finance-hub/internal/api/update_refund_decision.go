package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

func (c *Client) UpdateRefundDecision(ctx context.Context, clientID int, refundID int, status string) error {
	var body bytes.Buffer

	err := json.NewEncoder(&body).Encode(shared.UpdateRefundStatus{
		Status: shared.ParseRefundStatus(status),
	})

	if err != nil {
		return err
	}

	url := fmt.Sprintf("/clients/%d/refunds/%d", clientID, refundID)
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
