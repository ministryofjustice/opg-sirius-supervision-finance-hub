package allpay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type CancelMandateRequest struct {
	ClosureDate time.Time
	ClientDetails
}

func (c *Client) CancelMandate(ctx context.Context, data *CancelMandateRequest) error {
	logger := c.logger(ctx)

	path := fmt.Sprintf("/Customers/%s/%s/%s/Mandates",
		c.schemeCode,
		base64.StdEncoding.EncodeToString([]byte(data.ClientReference)),
		base64.StdEncoding.EncodeToString([]byte(trimChars(data.Surname, 19))),
	)

	today := time.Now().UTC().Truncate(24 * time.Hour)
	// closure date must be in the future, so if it is today's date, we omit it from the request
	if data.ClosureDate.UTC().Truncate(24 * time.Hour).After(today) {
		path = fmt.Sprintf("%s/%s", path, data.ClosureDate.Format("2006-01-02"))
	}

	req, err := c.newRequest(ctx, http.MethodDelete, path, nil)

	if err != nil {
		logger.Error("unable to build cancel mandate request", "error", err)
		return apiError("Direct Debit cannot be cancelled due to an unexpected system error.")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		logger.Error("unable to send cancel mandate request", "error", err)
		return apiError("Direct Debit cannot be cancelled due to an unexpected system error.")
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode == http.StatusUnprocessableEntity {
		var ve ErrorValidation

		err = json.NewDecoder(resp.Body).Decode(&ve)
		if err != nil {
			logger.Error("unable to parse cancel mandate validation response", "error", err)
			return apiError("Direct Debit cannot be cancelled due to an unexpected response from AllPay.")
		}

		if isAlreadyCancelledValidationError(ve) {
			logger.Info("mandate already cancelled in Allpay, treating as success", "messages", ve.Messages)
			return nil
		}

		logger.Error("cancel mandate request returned validation errors", "errors", ve)
		return ve
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("cancel mandate request returned unexpected status code", "status", resp.Status)
		return apiError("Direct Debit cannot be cancelled due to an unexpected response from AllPay.")
	}

	return nil
}

func isAlreadyCancelledValidationError(err ErrorValidation) bool {
	for _, message := range err.Messages {
		formattedMessage := strings.ToLower(message)
		if strings.Contains(formattedMessage, "a direct debit mandate was not found for this account") {
			return true
		}
	}

	return false
}
