package allpay

import (
	"context"
	"io"
	"net/http"
)

type Envs struct {
	apiHost    string
	apiKey     string
	schemeCode string
}

type ClientDetails struct {
	ClientReference string
	Surname         string
}

type Client struct {
	http HTTPClient
	Envs
}

func NewClient(httpClient HTTPClient, apiHost string, apiKey string, schemeCode string) *Client {
	return &Client{
		http: httpClient,
		Envs: Envs{
			apiHost:    apiHost,
			apiKey:     apiKey,
			schemeCode: schemeCode,
		},
	}
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.apiHost+"/AllpayApi"+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.apiKey)
	req.Header.Add("Content-Type", "application/json")

	return req, err
}

// unchecked allows errors to be unchecked when deferring a function, e.g. closing a reader where a failure would only
// occur when the process is likely to already be unrecoverable
func unchecked(f func() error) {
	_ = f()
}
