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

type schedule struct {
	Date   string `json:"ScheduleDate"`
	Amount int32  `json:"Amount"`
}

type createScheduleRequest struct {
	Schedules []schedule `json:"Schedules"`
}

type CreateScheduleInput struct {
	Date   time.Time
	Amount int32
	ClientDetails
}

func (c *Client) CreateSchedule(ctx context.Context, data *CreateScheduleInput) error {
	logger := c.logger(ctx)

	var body bytes.Buffer

	s := createScheduleRequest{
		Schedules: []schedule{{
			Date:   data.Date.Format("2006-01-02"),
			Amount: data.Amount,
		}},
	}

	err := json.NewEncoder(&body).Encode(s)
	if err != nil {
		logger.Error("unable to parse create schedule request", "error", err)
		return err
	}

	req, err := c.newRequest(ctx, http.MethodPost,
		fmt.Sprintf("/Customers/%s/%s/%s/VariableMandates",
			c.schemeCode,
			base64.StdEncoding.EncodeToString([]byte(data.ClientReference)),
			base64.StdEncoding.EncodeToString([]byte(data.Surname)),
		), &body)

	if err != nil {
		logger.Error("unable to build create schedule request", "error", err)
		return ErrorAPI{}
	}

	logger.Info("sending create schedule request", "url", req.URL.String(), "query", req.URL.RawQuery)

	resp, err := c.http.Do(req)
	if err != nil {
		logger.Error("unable to send create schedule request", "error", err)
		return ErrorAPI{}
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode == http.StatusUnprocessableEntity {
		var ve ErrorValidation

		err = json.NewDecoder(resp.Body).Decode(&ve)
		if err != nil {
			logger.Error("unable to parse create schedule validation response", "error", err)
			return ErrorAPI{}
		}

		logger.Error("create schedule request returned validation errors", "errors", ve)
		return ve
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("create schedule request returned unexpected status code", "status", resp.Status)
		return ErrorAPI{}
	}

	return nil
}
