package api

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/auth"
	"io"
	"net/http"
)

const ErrUnauthorized ClientError = "unauthorized"

type ClientError string

func (e ClientError) Error() string {
	return string(e)
}

type Envs struct {
	SiriusURL     string
	BackendURL    string
	HolidayAPIURL string
}

type JWTClient interface {
	CreateJWT(ctx context.Context) string
}

type AllPay interface {
	CancelMandate(ctx context.Context, data *allpay.CancelMandateRequest) error
	CreateMandate(ctx context.Context, data *allpay.CreateMandateRequest) error
	ModulusCheck(ctx context.Context, sortCode string, accountNumber string) error
	CreateSchedule(ctx context.Context, data allpay.CreateScheduleInput) error
}

type Client struct {
	http         HTTPClient
	caches       *Caches
	jwt          JWTClient
	allpayClient AllPay
	Envs
}

func NewClient(httpClient HTTPClient, jwt JWTClient, envs Envs, allpayClient AllPay) *Client {
	return &Client{
		http:         httpClient,
		caches:       newCaches(),
		jwt:          jwt,
		allpayClient: allpayClient,
		Envs: Envs{
			SiriusURL:     envs.SiriusURL,
			BackendURL:    envs.BackendURL,
			HolidayAPIURL: envs.HolidayAPIURL,
		},
	}
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (c *Client) newSiriusRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.SiriusURL+"/supervision-api/v1"+path, body)
	if err != nil {
		return nil, err
	}

	addCookiesFromContext(ctx, req)
	addXsrfFromContext(ctx, req)
	req.Header.Add("OPG-Bypass-Membrane", "1")

	return req, err
}

func (c *Client) newBackendRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.BackendURL+path, body)
	if err != nil {
		return nil, err
	}

	addCookiesFromContext(ctx, req)
	req.Header.Add("Authorization", "Bearer "+c.jwt.CreateJWT(ctx))

	return req, err
}

func (c *Client) newSessionRequest(ctx context.Context) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.SiriusURL+"/supervision-api/v1/users/current", nil)
	if err != nil {
		return nil, err
	}

	addCookiesFromContext(ctx, req)
	req.Header.Add("OPG-Bypass-Membrane", "1")

	return req, err
}

func (c *Client) newHolidayRequest(ctx context.Context) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.HolidayAPIURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")

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

// unchecked allows errors to be unchecked when deferring a function, e.g. closing a reader where a failure would only
// occur when the process is likely to already be unrecoverable
func unchecked(f func() error) {
	_ = f()
}
