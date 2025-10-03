package allpay

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/telemetry"
)

type ScheduleComponent struct {
	CollectionDate time.Time
	Amount         int32
}

type RemoveScheduleComponentRequests struct {
	ScheduleComponents []ScheduleComponent
	ClientDetails
}

func (c *Client) RemoveScheduleComponents(ctx context.Context, data *RemoveScheduleComponentRequests) error {
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

func (c *Client) removeScheduleComponent(ctx context.Context, clientDetails ClientDetails, schedule *ScheduleComponent) error {
	logger := telemetry.LoggerFromContext(ctx)

	req, err := c.newRequest(ctx, http.MethodDelete,
		fmt.Sprintf("/Customers/%s/%s/%s/Mandates/Schedule/%s/%s",
			c.schemeCode,
			base64.StdEncoding.EncodeToString([]byte(clientDetails.ClientReference)),
			base64.StdEncoding.EncodeToString([]byte(clientDetails.Surname)),
			schedule.CollectionDate.Format("2006-01-02"),
			schedule.Amount,
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
