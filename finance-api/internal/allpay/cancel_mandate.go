package allpay

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/telemetry"
)

func (c *Client) CancelMandate(ctx context.Context, data *ClientDetails) error {
	logger := telemetry.LoggerFromContext(ctx)

	today := time.Now().Format("2006-01-02")

	req, err := c.newRequest(ctx, http.MethodDelete,
		fmt.Sprintf("/Customers/%s/%s/%s/Mandates/%s",
			c.schemeCode,
			base64.StdEncoding.EncodeToString([]byte(data.ClientReference)),
			base64.StdEncoding.EncodeToString([]byte(data.Surname)),
			today,
		), nil)

	if err != nil {
		logger.Error("unable to build cancel mandate request", "error", err)
		return ErrorAPI{}
	}

	resp, err := c.http.Do(req)
	if err != nil {
		logger.Error("unable to send cancel mandate request", "error", err)
		return ErrorAPI{}
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode != http.StatusOK {
		logger.Error("cancel mandate request returned unexpected status code", "status", resp.Status)
		return ErrorAPI{}
	}

	return nil
}
