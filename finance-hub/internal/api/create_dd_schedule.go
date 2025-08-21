package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
	"time"
)

func (c *Client) CreateDirectDebitSchedule(ctx context.Context, clientId int) error {
	var body bytes.Buffer
	logger := telemetry.LoggerFromContext(ctx)

	balance, err := c.getPendingOutstandingBalance(ctx, clientId)
	if err != nil {
		logger.Error("failed to create schedule due to error in fetching outstanding balance", "error", err)
		return err
	}
	if balance < 1 {
		logger.Info(fmt.Sprintf("skipping direct debit schedule creation for client %d due to lack of outstanding balance", clientId), "balance", balance)
		return nil
	}

	client, err := c.GetPersonDetails(ctx, clientId)
	if err != nil {
		logger.Error("failed to create schedule due to error in fetching client details", "error", err)
		return err
	}

	date, err := c.addWorkingDays(ctx, time.Now().UTC(), 14)
	if err != nil {
		logger.Error("failed to create schedule due to error in calculating working days", "error", err)
		return err
	}

	date, _ = c.lastWorkingDayOfMonth(ctx, date) // no need to check error here as it would have failed earlier

	schedule := allpay.CreateScheduleInput{
		ClientRef: client.CourtRef,
		Surname:   client.Surname,
		Date:      date,
		Amount:    balance,
	}

	err = json.NewEncoder(&body).Encode(schedule)
	if err != nil {
		return err
	}

	err = c.allpayClient.CreateSchedule(ctx, schedule)
	if err != nil {
		var ve allpay.ErrorValidation
		if errors.As(err, &ve) {
			// we validate in advance so validation errors from AllPay should never occur
			// if they do, log them so we can investigate
			logger.Error("validation errors returned from allpay", "errors", ve.Messages)
		}
		return err
	}

	err = c.recordPendingCollection(ctx, clientId, shared.PendingCollection{
		Amount:         balance,
		CollectionDate: shared.Date{Time: date},
	})
	if err != nil {
		logger.Error("failed to create pending collection in Sirius after successful schedule instruction in AllPay", "error", err)
		return err
	}
	return nil
}

func (c *Client) getPendingOutstandingBalance(ctx context.Context, clientId int) (int, error) {
	var v int
	req, err := c.newBackendRequest(ctx, http.MethodGet, fmt.Sprintf("/clients/%d/balance/pending", clientId), nil)
	if err != nil {
		return v, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return v, err
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode == http.StatusUnauthorized {
		return v, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return v, newStatusError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(&v)
	return v, err
}

func (c *Client) recordPendingCollection(ctx context.Context, clientId int, data shared.PendingCollection) error {
	var body bytes.Buffer

	err := json.NewEncoder(&body).Encode(data)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("/clients/%d/pending-collections", clientId)
	req, err := c.newBackendRequest(ctx, http.MethodPost, url, &body)

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

	return newStatusError(resp)
}
