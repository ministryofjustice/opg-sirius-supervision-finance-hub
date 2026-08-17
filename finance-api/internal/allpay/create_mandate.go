package allpay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

type Customer struct {
	ClientDetails
	Address    Address `json:"Address"`
	SchemeCode string  `json:"SchemeCode"`
}

type Schedule struct {
	Date          string `json:"ScheduleDate"`
	Amount        int32  `json:"Amount"`
	Frequency     string `json:"Frequency"`
	TotalPayments int32  `json:"TotalPayments"`
}

type CreateMandateRequest struct {
	Customer    Customer `json:"Customer"`
	BankAccount struct {
		BankDetails BankDetails `json:"BankDetails"`
	} `json:"BankAccount"`
	Schedules []Schedule `json:"Schedules,omitempty"`
}

type CreateMandateInput struct {
	Customer    Customer
	BankDetails BankDetails
	Schedule    *ScheduleInput
}

func (c *Client) CreateMandate(ctx context.Context, input *CreateMandateInput) error {
	logger := c.logger(ctx)

	data := CreateMandateRequest{
		Customer: Customer{
			ClientDetails: input.Customer.ClientDetails,
			Address: Address{
				Line1:    trimChars(input.Customer.Address.Line1, 40),
				Town:     trimChars(input.Customer.Address.Town, 40),
				PostCode: trimChars(input.Customer.Address.PostCode, 8),
			},
			SchemeCode: c.schemeCode, // add scheme code here instead of leaking it outside the client
		},
		BankAccount: struct {
			BankDetails BankDetails `json:"BankDetails"`
		}{
			BankDetails: input.BankDetails,
		},
	}

	if input.Schedule != nil {
		data.Schedules = []Schedule{input.Schedule.ToSchedule()}
	}

	var body bytes.Buffer

	err := json.NewEncoder(&body).Encode(data)
	if err != nil {
		logger.Error("unable to parse create mandate request", "error", err)
		return err
	}

	req, err := c.newRequest(ctx, http.MethodPost, fmt.Sprintf("/Customers/%s/Mandates/Create", c.schemeCode), &body)

	if err != nil {
		logger.Error("unable to build create mandate request", "error", err)
		return apiError("Direct Debit cannot be setup due to an unexpected system error.")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		logger.Error("unable to send create mandate request", "error", err)
		return apiError("Direct Debit cannot be setup due to an unexpected system error.")
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode == http.StatusUnprocessableEntity {
		var ve ErrorValidation

		err = json.NewDecoder(resp.Body).Decode(&ve)
		if err != nil {
			logger.Error("unable to parse create mandate validation response", "error", err)
			return apiError("Direct Debit cannot be setup due to an unexpected response from AllPay.")
		}

		if alreadyExistsValidationError(ve) {
			logger.Info("mandate already exists in Allpay", "messages", ve.Messages)

			if input.Schedule != nil {
				logger.Info("creating schedule for existing mandate in Allpay")
				return c.CreateSchedule(ctx, &CreateScheduleInput{
					ScheduleInput: *input.Schedule,
					ClientDetails: input.Customer.ClientDetails,
				})
			}
			return nil
		}

		logger.Error("create mandate request returned validation errors", "errors", ve)
		return ve
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("create mandate request returned unexpected status code", "status", resp.Status)
		return apiError("Direct Debit cannot be setup due to an unexpected response from AllPay.")
	}

	return nil
}

func alreadyExistsValidationError(err ErrorValidation) bool {
	for _, message := range err.Messages {
		formattedMessage := strings.ToLower(message)
		if strings.Contains(formattedMessage, "this direct debit reference is already being used") {
			return true
		}
	}

	return false
}
