package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (c *Client) CancelDirectDebitMandate(ctx context.Context, clientId int) error {
	var body bytes.Buffer
	logger := telemetry.LoggerFromContext(ctx)

	client, err := c.GetPersonDetails(ctx, clientId)
	if err != nil {
		return err
	}

	err = json.NewEncoder(&body).Encode(&shared.CancelMandate{
		AllPayCustomer: shared.AllPayCustomer{
			ClientReference: client.CourtRef,
			Surname:         client.Surname,
		},
	})
	if err != nil {
		return err
	}

	req, err := c.newBackendRequest(ctx, http.MethodDelete, fmt.Sprintf("/clients/%d/direct-debit", clientId), &body)

	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode != http.StatusNoContent {
		logger.Error("cancel mandate request returned unexpected status code", "status", resp.Status)
		return newStatusError(resp)
	}

	return nil
}
