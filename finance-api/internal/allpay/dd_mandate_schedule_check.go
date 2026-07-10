package allpay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type FetchMandateScheduleInput struct {
	ClientDetails
}

type FetchMandateScheduleDataType []FetchMandateScheduleDataRecord

type FetchMandateScheduleDataRecord struct {
	Amount          int32  `json:"Amount"`
	ClientReference string `json:"ClientReference"`
	LastName        string `json:"LastName"`
	ScheduleDate    string `json:"ScheduleDate"`
	Status          string `json:"Status"`
}

type FetchMandateScheduleOutput struct {
	FetchMandateScheduleDataType FetchMandateScheduleDataType `json:"FetchMandateScheduleData"`
	TotalRecords                 int                          `json:"TotalRecords"`
}

type MandateScheduleCheckOutput struct {
	Mandate       *FetchMandateScheduleOutput
	Schedule      *FetchMandateScheduleOutput
	MandateError  string
	ScheduleError string
}

func (c *Client) FetchMandate(ctx context.Context, data FetchMandateScheduleInput) (*FetchMandateScheduleOutput, error) {
	return c.fetchMandateScheduleData(ctx, data, "", "mandate")
}

func (c *Client) FetchSchedule(ctx context.Context, data FetchMandateScheduleInput) (*FetchMandateScheduleOutput, error) {
	return c.fetchMandateScheduleData(ctx, data, "/Mandates/Schedule", "schedule")
}

func (c *Client) fetchMandateScheduleData(ctx context.Context, data FetchMandateScheduleInput, suffix string, resource string) (*FetchMandateScheduleOutput, error) {
	logger := c.logger(ctx)

	req, err := c.newRequest(ctx, http.MethodGet,
		fmt.Sprintf("/Customers/%s/%s/%s%s",
			c.schemeCode,
			base64.StdEncoding.EncodeToString([]byte(data.ClientReference)),
			base64.StdEncoding.EncodeToString([]byte(trimChars(data.Surname, 19))),
			suffix,
		), nil)
	if err != nil {
		logger.Error(fmt.Sprintf("unable to build %s fetch request", resource), "error", err)
		return nil, apiError(fmt.Sprintf("%s data cannot be fetched due to an unexpected system error.", resource))
	}

	resp, err := c.http.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("unable to send %s fetch request", resource), "error", err)
		return nil, apiError(fmt.Sprintf("%s data cannot be fetched due to an unexpected system error.", resource))
	}
	defer unchecked(resp.Body.Close)

	if resp.StatusCode != http.StatusOK {
		logger.Error(fmt.Sprintf("%s fetch request returned unexpected status code", resource), "status", resp.Status)
		return nil, apiError(fmt.Sprintf("%s data cannot be fetched due to an unexpected response from AllPay.", resource))
	}

	var body FetchMandateScheduleOutput
	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		logger.Error(fmt.Sprintf("unable to parse %s fetch response", resource), "error", err)
		return nil, apiError(fmt.Sprintf("%s data cannot be fetched due to an unexpected response from AllPay.", resource))
	}

	return &body, nil
}
