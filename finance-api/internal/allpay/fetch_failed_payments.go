package allpay

import (
	"context"
	"time"
)

type FetchFailedPaymentsRequest struct {
	From time.Time
	To   time.Time
}

func (c *Client) FetchFailedPayments(ctx context.Context, data *FetchFailedPaymentsRequest) error {
	//logger := telemetry.LoggerFromContext(ctx)
	//
	//today := time.Now().Format("2006-01-02")
	//
	//req, err := c.newRequest(ctx, http.MethodDelete, fmt.Sprintf("/Customers/%s/%s/%s/Mandates/%s", c.schemeCode, data.ClientReference, data.Surname, today), nil)
	//
	//if err != nil {
	//	logger.Error("unable to build cancel mandate request", "error", err)
	//	return ErrorAPI{}
	//}
	//
	//resp, err := c.http.Do(req)
	//if err != nil {
	//	logger.Error("unable to send cancel mandate request", "error", err)
	//	return ErrorAPI{}
	//}
	//
	//defer unchecked(resp.Body.Close)
	//
	//if resp.StatusCode != http.StatusOK {
	//	logger.Error("cancel mandate request returned unexpected status code", "status", resp.Status)
	//	return ErrorAPI{}
	//}

	return nil
}
