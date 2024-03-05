package api

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-go-common/logging"
	"io"
	"net/http"
)

const ErrUnauthorized ClientError = "unauthorized"

type ClientError string

func (e ClientError) Error() string {
	return string(e)
}

type ValidationErrors map[string]map[string]string

type ValidationError struct {
	Message string
	Errors  ValidationErrors
}

func (ve ValidationError) Error() string {
	return ve.Message
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

func (e StatusError) Title() string {
	return "unexpected response from Sirius"
}

func (e StatusError) Data() interface{} {
	return e
}

type Context struct {
	Context   context.Context
	Cookies   []*http.Cookie
	XSRFToken string
	ClientId  int
}

func (ctx Context) With(c context.Context) Context {
	return Context{
		Context:   c,
		Cookies:   ctx.Cookies,
		XSRFToken: ctx.XSRFToken,
	}
}

func NewApiClient(httpClient HTTPClient, siriusUrl string, backendUrl string, logger *logging.Logger) (*ApiClient, error) {
	return &ApiClient{
		http:       httpClient,
		siriusUrl:  siriusUrl,
		backendUrl: backendUrl,
		logger:     logger,
	}, nil
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ApiClient struct {
	http       HTTPClient
	siriusUrl  string
	logger     *logging.Logger
	backendUrl string
}

func (c *ApiClient) newSiriusRequest(ctx Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx.Context, method, c.siriusUrl+path, body)
	if err != nil {
		return nil, err
	}

	for _, c := range ctx.Cookies {
		req.AddCookie(c)
	}

	req.Header.Add("OPG-Bypass-Membrane", "1")
	req.Header.Add("X-XSRF-TOKEN", ctx.XSRFToken)

	return req, err
}

func (c *ApiClient) newBackendRequest(ctx Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx.Context, method, c.backendUrl+path, body)
	if err != nil {
		return nil, err
	}

	for _, c := range ctx.Cookies {
		req.AddCookie(c)
	}

	req.Header.Add("OPG-Bypass-Membrane", "1")
	req.Header.Add("X-XSRF-TOKEN", ctx.XSRFToken)

	return req, err
}

func (c *ApiClient) logErrorRequest(req *http.Request, err error) {
	c.logger.Print("method: " + req.Method + ", url: " + req.URL.Path)
	if err != nil {
		c.logger.Print(err)
	}
}
