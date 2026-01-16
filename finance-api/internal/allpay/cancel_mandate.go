package allpay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/telemetry"
)

type CancelMandateRequest struct {
	ClosureDate time.Time
	ClientDetails
}

func (c *Client) CancelMandate(ctx context.Context, data *CancelMandateRequest) error {
	logger := telemetry.LoggerFromContext(ctx)

	req, err := c.newRequest(ctx, http.MethodDelete,
		fmt.Sprintf("/Customers/%s/%s/%s/Mandates/%s",
			c.schemeCode,
			base64.StdEncoding.EncodeToString([]byte(data.ClientReference)),
			base64.StdEncoding.EncodeToString([]byte(data.Surname)),
			data.ClosureDate.Format("2006-01-02"),
		), nil)

	if err != nil {
		logger.Error("unable to build cancel mandate request", "error", err)
		return ErrorAPI{}
	}

	logger.Info("sending cancel mandate request", "url", req.URL.String(), "query", req.URL.RawQuery)

	resp, err := c.http.Do(req)
	if err != nil {
		logger.Error("unable to send cancel mandate request", "error", err)
		return ErrorAPI{}
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode == http.StatusUnprocessableEntity {
		var ve ErrorValidation

		err = json.NewDecoder(resp.Body).Decode(&ve)
		if err != nil {
			logger.Error("unable to parse cancel mandate validation response", "error", err)
			return ErrorAPI{}
		}

		logger.Error("cancel mandate request returned validation errors", "errors", ve)
		return ve
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("cancel mandate request returned unexpected status code", "status", resp.Status)
		return ErrorAPI{}
	}

	return nil
}
