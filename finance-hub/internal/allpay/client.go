package allpay

import (
	"context"
	"io"
	"net/http"
)

type Envs struct {
	APIHost    string
	APIKey     string
	SchemeCode string
}

type Client struct {
	http HTTPClient
	Envs
}

func NewClient(httpClient HTTPClient, envs Envs) *Client {
	return &Client{
		http: httpClient,
		Envs: Envs{
			APIHost:    envs.APIHost,
			APIKey:     envs.APIKey,
			SchemeCode: "OPGB",
		},
	}
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.APIHost+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.APIKey)
	req.Header.Add("Accept", "application/json")

	return req, err
}

// unchecked allows errors to be unchecked when deferring a function, e.g. closing a reader where a failure would only
// occur when the process is likely to already be unrecoverable
func unchecked(f func() error) {
	_ = f()
}
