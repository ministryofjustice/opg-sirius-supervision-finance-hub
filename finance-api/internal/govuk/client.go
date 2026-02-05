package govuk

import (
	"context"
	"net/http"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Envs struct {
	HolidayAPIURL string
}
type Client struct {
	http   HTTPClient
	caches *Caches
	Envs
}

func NewClient(httpClient HTTPClient, holidayApi string) *Client {
	return &Client{
		http:   httpClient,
		caches: newCaches(),
		Envs:   Envs{HolidayAPIURL: holidayApi},
	}
}

func (c *Client) newHolidayRequest(ctx context.Context) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.HolidayAPIURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")

	return req, err
}

// unchecked allows errors to be unchecked when deferring a function, e.g. closing a reader where a failure would only
// occur when the process is likely to already be unrecoverable
func unchecked(f func() error) {
	_ = f()
}
