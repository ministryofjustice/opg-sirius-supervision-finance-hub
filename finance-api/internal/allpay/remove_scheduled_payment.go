package allpay

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/telemetry"
)

type RemoveScheduledPaymentRequest struct {
	CollectionDate time.Time
	Amount         int32
	ClientDetails
}

func (c *Client) RemoveScheduledPayment(ctx context.Context, data *RemoveScheduledPaymentRequest) error {
	logger := telemetry.LoggerFromContext(ctx)

	req, err := c.newRequest(ctx, http.MethodDelete,
		fmt.Sprintf("/Customers/%s/%s/%s/Mandates/Schedule/%s/%d",
			c.schemeCode,
			base64.StdEncoding.EncodeToString([]byte(data.ClientReference)),
			base64.StdEncoding.EncodeToString([]byte(data.Surname)),
			data.CollectionDate.Format("2006-01-02"),
			data.Amount,
		), nil)

	if err != nil {
		logger.Error("unable to build remove schedule component request", "error", err)
		return ErrorAPI{}
	}

	resp, err := c.http.Do(req)
	if err != nil {
		logger.Error("unable to send remove schedule component request", "error", err)
		return ErrorAPI{}
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode != http.StatusOK {
		logger.Error("remove schedule component request returned unexpected status code", "status", resp.Status)
		return ErrorAPI{}
	}

	return nil
}
