package api

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/auth"
	"io"
	"net/http"
)

const ErrUnauthorized ClientError = "unauthorized"

type ClientError string

func (e ClientError) Error() string {
	return string(e)
}

func NewApiClient(httpClient HTTPClient, siriusUrl string, backendUrl string) (*Client, error) {
	return &Client{
		http:       httpClient,
		siriusUrl:  siriusUrl,
		backendUrl: backendUrl,
		caches:     newCaches(),
	}, nil
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	http       HTTPClient
	siriusUrl  string
	backendUrl string
	caches     *Caches
}

func (c *Client) newSiriusRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.siriusUrl+"/supervision-api/v1"+path, body)
	if err != nil {
		return nil, err
	}

	addCookiesFromContext(ctx, req)
	addXsrfFromContext(ctx, req)
	req.Header.Add("OPG-Bypass-Membrane", "1")

	return req, err
}

func (c *Client) newBackendRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.backendUrl+path, body)
	if err != nil {
		return nil, err
	}

	addCookiesFromContext(ctx, req)

	return req, err
}

func addCookiesFromContext(ctx context.Context, req *http.Request) {
	for _, c := range ctx.(auth.Context).Cookies {
		req.AddCookie(c)
	}
}

func addXsrfFromContext(ctx context.Context, req *http.Request) {
	req.Header.Add("X-XSRF-TOKEN", ctx.(auth.Context).XSRFToken)
}

type StatusError struct {
	Code   int    `json:"code"`
	URL    string `json:"url"`
	Method string `json:"method"`
}

func newStatusError(resp *http.Response) StatusError {
	return StatusError{
		Code:   resp.StatusCode,
		URL:    resp.Request.URL.String(),
		Method: resp.Request.Method,
	}
}

func (e StatusError) Error() string {
	return fmt.Sprintf("%s %s returned %d", e.Method, e.URL, e.Code)
}

func (e StatusError) Data() interface{} {
	return e
}
