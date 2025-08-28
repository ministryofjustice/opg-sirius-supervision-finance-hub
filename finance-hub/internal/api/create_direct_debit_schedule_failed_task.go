package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (c *Client) CreateDirectDebitScheduleFailedTask(ctx context.Context, clientId int) error {
	var body bytes.Buffer

	task := shared.Task{
		ClientId: clientId,
		Type:     "FDSC",
		DueDate:  shared.Date{Time: time.Now()},
		Notes:    fmt.Sprintf("The creation of a direct debit collection schedule for this client has failed"),
	}

	err := json.NewEncoder(&body).Encode(task)
	if err != nil {
		return err
	}

	req, err := c.newSiriusRequest(ctx, http.MethodPost, "/tasks", &body)

	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode == http.StatusCreated {
		return nil
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}
	if resp.StatusCode == http.StatusUnprocessableEntity {
		var v apierror.ValidationError
		if err = json.NewDecoder(resp.Body).Decode(&v); err == nil && len(v.Errors) > 0 {
			return apierror.ValidationError{Errors: v.Errors}
		}
	}
	if resp.StatusCode == http.StatusBadRequest {
		var be apierror.BadRequest
		if err = json.NewDecoder(resp.Body).Decode(&be); err == nil {
			return be
		}
	}

	return newStatusError(resp)
}
