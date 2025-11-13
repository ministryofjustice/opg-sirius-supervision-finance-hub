package allpay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/telemetry"
)

type modulusCheckResponse struct {
	Valid              bool `json:"Valid"`
	DirectDebitCapable bool `json:"DirectDebitCapable"`
}

func (c *Client) ModulusCheck(ctx context.Context, sortCode string, accountNumber string) error {
	logger := telemetry.LoggerFromContext(ctx)
	url := fmt.Sprintf("/BankAccounts?sortcode=%s&accountnumber=%s", sortCode, accountNumber)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiHost+"/AllpayApi"+url, nil)

	if err != nil {
		logger.Error("unable to build modulus check request", "error", err)
		return ErrorAPI{}
	}

	req.Header.Add("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		logger.Error("unable to send modulus check request", "error", err)
		return ErrorAPI{}
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode != http.StatusOK {
		logger.Error("modulus check request returned unexpected status code", "status", resp.Status)
		return ErrorAPI{}
	}

	var modulusCheck modulusCheckResponse

	err = json.NewDecoder(resp.Body).Decode(&modulusCheck)
	if err != nil {
		logger.Error("unable to parse modulus check response", "error", err)
		return ErrorAPI{}
	}

	if !modulusCheck.Valid || !modulusCheck.DirectDebitCapable {
		return ErrorModulusCheckFailed{}
	}

	return nil
}
