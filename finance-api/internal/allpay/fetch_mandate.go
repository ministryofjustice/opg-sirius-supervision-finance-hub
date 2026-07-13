package allpay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type FetchMandateInput struct {
	ClientDetails
}

type FetchMandateData []FetchMandateDataRecord

type FetchMandateDataRecord struct {
	ClientReference string `json:"ClientReference"`
	LastName        string `json:"LastName"`
	Status          string `json:"Status"`
}

type FetchMandateOutput struct {
	FetchMandateData FetchMandateData `json:"FetchMandateScheduleData"`
	MandateError     string
	TotalRecords     int `json:"TotalRecords"`
}

func (c *Client) FetchMandate(ctx context.Context, input FetchMandateInput) (*FetchMandateOutput, error) {

	logger := c.logger(ctx)

	req, err := c.newRequest(ctx, http.MethodGet,
		fmt.Sprintf("/Customers/%s/%s/%s",
			c.schemeCode,
			base64.StdEncoding.EncodeToString([]byte(input.ClientReference)),
			base64.StdEncoding.EncodeToString([]byte(trimChars(input.Surname, 19)))), nil)

	if err != nil {
		logger.Error("unable to build mandate fetch request", "error", err)
		return nil, apiError("mandate data cannot be fetched due to an unexpected system error.")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		logger.Error("unable to send mandate fetch request", "error", err)
		return nil, apiError("mandate data cannot be fetched due to an unexpected system error.")
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode != http.StatusOK {
		logger.Error("mandate fetch request returned unexpected status code", "status", resp.Status)
		return nil, apiError("mandate data cannot be fetched due to an unexpected response from AllPay.")
	}

	var body FetchMandateOutput

	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		logger.Error("unable to parse mandate fetch response", "error", err)
		return nil, apiError("mandate data cannot be fetched due to an unexpected response from AllPay.")
	}

	return &body, nil
}
