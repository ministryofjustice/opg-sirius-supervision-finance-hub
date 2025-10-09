package allpay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/telemetry"
)

type BankDetails struct {
	AccountName   string `json:"AccountName"`
	SortCode      string `json:"SortCode"`
	AccountNumber string `json:"AccountNumber"`
}

type Address struct {
	Line1    string `json:"Line1"`
	Town     string `json:"Town"`
	PostCode string `json:"PostCode"`
}

type CreateMandateRequest struct {
	SchemeCode      string  `json:"SchemeCode"`
	ClientReference string  `json:"ClientReference"`
	Surname         string  `json:"LastName"`
	Address         Address `json:"Address"`
	BankAccount     struct {
		BankDetails BankDetails `json:"BankDetails"`
	} `json:"BankAccount"`
}

func (c *Client) CreateMandate(ctx context.Context, data *CreateMandateRequest) error {
	logger := telemetry.LoggerFromContext(ctx)

	// add scheme code here instead of leaking it outside the client
	data.SchemeCode = c.schemeCode

	var body bytes.Buffer

	err := json.NewEncoder(&body).Encode(data)
	if err != nil {
		logger.Error("unable to parse create mandate request", "error", err)
		return err
	}

	req, err := c.newRequest(ctx, http.MethodPost, fmt.Sprintf("/Customers/%s/VariableMandates/Create", c.schemeCode), &body)

	if err != nil {
		logger.Error("unable to build create mandate request", "error", err)
		return ErrorAPI{}
	}

	resp, err := c.http.Do(req)
	if err != nil {
		logger.Error("unable to send create mandate request", "error", err)
		return ErrorAPI{}
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode == http.StatusUnprocessableEntity {
		var ve ErrorValidation

		err = json.NewDecoder(resp.Body).Decode(&ve)
		if err != nil {
			logger.Error("unable to parse create mandate validation response", "error", err)
			return ErrorAPI{}
		}

		logger.Error("create mandate request returned validation errors", "errors", ve)
		return ve
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("create mandate request returned unexpected status code", "status", resp.Status)
		return ErrorAPI{}
	}

	return nil
}
