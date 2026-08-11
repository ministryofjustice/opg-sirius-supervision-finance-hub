package allpay

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type createScheduleRequest struct {
	Schedules []Schedule `json:"Schedules"`
}

type ScheduleInput struct {
	Date   time.Time
	Amount int32
}

func (s ScheduleInput) ToSchedule() Schedule {
	return Schedule{
		Date:          s.Date.Format("2006-01-02"),
		Amount:        s.Amount,
		Frequency:     "1",
		TotalPayments: 1,
	}
}

type CreateScheduleInput struct {
	ScheduleInput
	ClientDetails
}

func (c *Client) CreateSchedule(ctx context.Context, data *CreateScheduleInput) error {
	logger := c.logger(ctx)

	var body bytes.Buffer

	s := createScheduleRequest{
		Schedules: []Schedule{data.ToSchedule()},
	}

	err := json.NewEncoder(&body).Encode(s)
	if err != nil {
		logger.Error("unable to parse create schedule request", "error", err)
		return err
	}

	req, err := c.newRequest(ctx, http.MethodPost,
		fmt.Sprintf("/Customers/%s/%s/%s/Mandates",
			c.schemeCode,
			base64.StdEncoding.EncodeToString([]byte(data.ClientReference)),
			base64.StdEncoding.EncodeToString([]byte(trimChars(data.Surname, 19))),
		), &body)

	if err != nil {
		logger.Error("unable to build create schedule request", "error", err)
		return apiError("Schedule cannot be created due to an unexpected system error.")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		logger.Error("unable to send create schedule request", "error", err)
		return apiError("Schedule cannot be created due to an unexpected system error.")
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode == http.StatusUnprocessableEntity {
		var ve ErrorValidation

		err = json.NewDecoder(resp.Body).Decode(&ve)
		if err != nil {
			logger.Error("unable to parse create schedule validation response", "error", err)
			return apiError("Schedule cannot be created due to an unexpected response from AllPay.")
		}

		logger.Error("create schedule request returned validation errors", "errors", ve)
		return ve
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("create schedule request returned unexpected status code", "status", resp.Status)
		return apiError("Schedule cannot be created due to an unexpected response from AllPay.")
	}

	return nil
}
