package allpay

import (
	"context"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockHttp struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

var (
	// GetDoFunc fetches the mock client's `Do` func. Implement this within a test to modify the client's behaviour.
	GetDoFunc func(req *http.Request) (*http.Response, error)
)

func TestNewRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer testkey", r.Header.Get("Authorization"))
	}))
	defer ts.Close()

	client := Client{
		http: ts.Client(),
		Envs: Envs{
			apiHost:    "",
			apiKey:     "testkey",
			schemeCode: "",
		},
	}
	_, _ = client.newRequest(context.Background(), "GET", "/", nil)
}

func testContext() context.Context {
	return telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("test"))
}
