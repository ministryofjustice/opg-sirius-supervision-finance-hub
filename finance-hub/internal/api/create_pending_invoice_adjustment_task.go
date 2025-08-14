package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
	"strings"
	"time"
)

func (c *Client) CreatePendingInvoiceAdjustmentTask(ctx context.Context, clientId int, supervisionBillingTeamId int, invoiceRef string, adjustmentType string) error {
	var body bytes.Buffer

	dueDate, _ := c.addWorkingDays(ctx, time.Now(), 20)
	adjustmentTypeLabel := strings.ToLower(strings.ReplaceAll(adjustmentType, "_", " "))

	task := shared.Task{
		ClientId: clientId,
		Type:     "FPIA",
		DueDate:  shared.Date{Time: dueDate},
		Assignee: supervisionBillingTeamId,
		Notes:    fmt.Sprintf("Pending %s added to %s requires manager approval", adjustmentTypeLabel, invoiceRef),
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
