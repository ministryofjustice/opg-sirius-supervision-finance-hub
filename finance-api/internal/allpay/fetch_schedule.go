package allpay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type FetchScheduleInput struct {
	ClientDetails
}

type FetchScheduleData []FetchScheduleDataRecord

type FetchScheduleDataRecord struct {
	Amount          int32  `json:"Amount"`
	ClientReference string `json:"ClientReference"`
	LastName        string `json:"LastName"`
	ScheduleDate    string `json:"ScheduleDate"`
	Status          string `json:"Status"`
}

type FetchScheduleOutput struct {
	FetchScheduleData FetchScheduleData `json:"FetchMandateScheduleData"`
	ScheduleError     string
	TotalRecords      int `json:"TotalRecords"`
}

func (c *Client) FetchSchedule(ctx context.Context, input FetchScheduleInput) (*FetchScheduleOutput, error) {
	logger := c.logger(ctx)

	req, err := c.newRequest(ctx, http.MethodGet,
		fmt.Sprintf("/Customers/%s/%s/%s/Mandates/Schedule",
			c.schemeCode,
			base64.StdEncoding.EncodeToString([]byte(input.ClientReference)),
			base64.StdEncoding.EncodeToString([]byte(trimChars(input.Surname, 19))),
		), nil)

	if err != nil {
		logger.Error("unable to build schedule fetch request", "error", err)
		return nil, apiError("schedule data cannot be fetched due to an unexpected system error.")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		logger.Error("unable to send schedule fetch request", "error", err)
		return nil, apiError("schedule data cannot be fetched due to an unexpected system error.")
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode != http.StatusOK {
		logger.Error("schedule fetch request returned unexpected status code", "status", resp.Status)
		return nil, apiError("schedule data cannot be fetched due to an unexpected response from AllPay.")
	}

	var body FetchScheduleOutput

	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		logger.Error("unable to parse schedule fetch response", "error", err)
		return nil, apiError("schedule data cannot be fetched due to an unexpected response from AllPay.")
	}

	return &body, nil
}
