package allpay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/telemetry"
)

type FetchFailedPaymentsInput struct {
	To   time.Time
	From time.Time
}

type FailedPayments []FailedPayment

type FailedPayment struct {
	Amount          int32  `json:"Amount"`
	ClientReference string `json:"ClientReference"`
	CollectionDate  string `json:"CollectionDate"`
	ProcessedDate   string `json:"ProcessedDate"`
	ReasonCode      string `json:"ReasonCode"`
}

type FailedPaymentsOutput struct {
	FailedPayments FailedPayments `json:"FailedPayments"`
	TotalRecords   int            `json:"TotalRecords"`
}

func (c *Client) FetchFailedPayments(ctx context.Context, input FetchFailedPaymentsInput) (FailedPayments, error) {
	fp := FailedPayments{}
	page := 1
	for {
		out, err := c.fetchFailedPaymentsForPage(ctx, input, page)
		if err != nil {
			return nil, err
		}
		if len(out.FailedPayments) == 0 {
			return fp, nil
		}
		fp = append(fp, out.FailedPayments...)
		page++
	}
}

func (c *Client) fetchFailedPaymentsForPage(ctx context.Context, input FetchFailedPaymentsInput, page int) (*FailedPaymentsOutput, error) {
	logger := telemetry.LoggerFromContext(ctx)

	req, err := c.newRequest(ctx, http.MethodGet, fmt.Sprintf("/Customers/%s/Mandates/FailedPayments/%s/%s/%d", c.schemeCode, input.From.Format("2006-01-02"), input.To.Format("2006-01-02"), page), nil)

	if err != nil {
		logger.Error("unable to build failed payments request", "error", err)
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		logger.Error("unable to send failed payments request", "error", err)
		return nil, err
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode != http.StatusOK {
		logger.Error("failed payments request returned unexpected status code", "status", resp.Status)
		return nil, ErrorAPI{}
	}

	var body FailedPaymentsOutput

	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		logger.Error("unable to parse failed payments response", "error", err)
		return nil, err
	}

	return &body, nil
}
