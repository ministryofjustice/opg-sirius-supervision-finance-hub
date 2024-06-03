package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
	"strings"
	"time"
)

func AddWorkingDays(date time.Time, days int) time.Time {
	for {
		if days == 0 {
			return date
		}

		date = date.AddDate(0, 0, 1)

		if date.Weekday() == time.Saturday {
			date = date.AddDate(0, 0, 2)
			return AddWorkingDays(date, days-1)
		} else if date.Weekday() == time.Sunday {
			date = date.AddDate(0, 0, 1)
			return AddWorkingDays(date, days-1)
		}

		days--
	}
}

func (c *ApiClient) CreatePendingInvoiceAdjustmentTask(ctx Context, clientId int, supervisionBillingTeamId int, invoiceId int, adjustmentType string) error {
	var body bytes.Buffer

	dueDate := AddWorkingDays(time.Now(), 20)
	adjustmentTypeLabel := strings.ToLower(strings.Replace(adjustmentType, "_", " ", -1))

	task := shared.Task{
		ClientId: clientId,
		Type:     "FPIA",
		DueDate:  dueDate.Format("02/01/2006"),
		Assignee: supervisionBillingTeamId,
		Notes:    fmt.Sprintf("Pending %s added to %d requires manager approval", adjustmentTypeLabel, invoiceId),
	}

	err := json.NewEncoder(&body).Encode(task)
	if err != nil {
		return err
	}

	req, err := c.newSiriusRequest(ctx, http.MethodPost, "/api/v1/tasks", &body)

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
		var v shared.ValidationError
		if err = json.NewDecoder(resp.Body).Decode(&v); err == nil && len(v.Errors) > 0 {
			return shared.ValidationError{Errors: v.Errors}
		}
	}
	if resp.StatusCode == http.StatusBadRequest {
		var be shared.BadRequest
		if err = json.NewDecoder(resp.Body).Decode(&be); err == nil {
			return be
		}
	}

	return newStatusError(resp)
}
