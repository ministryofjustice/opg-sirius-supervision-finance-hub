package allpay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"net/http"
)

var (
	ErrorModulusCheckFailed = errors.New("modulus check failed")
)

type modulusCheckResponse struct {
	Valid              bool `json:"Valid"`
	DirectDebitCapable bool `json:"DirectDebitCapable"`
}

func (c *Client) ModulusCheck(ctx context.Context, sortCode string, accountNumber string) (bool, error) {
	logger := telemetry.LoggerFromContext(ctx)
	req, err := c.newRequest(ctx, http.MethodGet, fmt.Sprintf("AllPayApi/BankAccounts?sortcode=%s&accountnumber=%s", sortCode, accountNumber), nil)

	if err != nil {
		logger.Error("unable to build modulus check request", "error", err)
		return false, ErrorAPI
	}

	resp, err := c.http.Do(req)
	if err != nil {
		logger.Error("unable to send modulus check request", "error", err)
		return false, ErrorAPI
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode != http.StatusOK {
		logger.Error("modulus check request returned unexpected status code", "status", resp.Status)
		return false, ErrorAPI
	}

	var modulusCheck modulusCheckResponse

	err = json.NewDecoder(resp.Body).Decode(&modulusCheck)
	if err != nil {
		logger.Error("unable to parse modulus check response", "error", err)
		return false, ErrorAPI
	}

	if !modulusCheck.Valid || !modulusCheck.DirectDebitCapable {
		return false, ErrorModulusCheckFailed
	}

	return true, nil
}
