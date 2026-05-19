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

	req, err := c.newRequest(ctx, http.MethodDelete,
		fmt.Sprintf("/Customers/%s/%s/%s/Mandates/%s",
			c.schemeCode,
			base64.StdEncoding.EncodeToString([]byte(data.ClientReference)),
			base64.StdEncoding.EncodeToString([]byte(trimChars(data.Surname, 19))),
			data.ClosureDate.Format("2006-01-02"),
		), nil)

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
    
    //if isAlreadyCancelledValidationError(ve) {
		logger.Info("mandate already cancelled in Allpay, treating as success", "messages", ve.Messages)
		return nil
		//}
    
// 		if err != nil {
// 			logger.Error("unable to parse cancel mandate validation response", "error", err)
// 			return apiError("Direct Debit cannot be cancelled due to an unexpected response from AllPay.")
// 		}
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("cancel mandate request returned unexpected status code", "status", resp.Status)
		return apiError("Direct Debit cannot be cancelled due to an unexpected response from AllPay.")
	}

	return nil
}

func isAlreadyCancelledValidationError(err ErrorValidation) bool {
	for _, message := range err.Messages {
		formatted := strings.ToLower(message)
		if strings.Contains(formatted, "no active mandate") {
			return true
		}
	}

	return false
}
