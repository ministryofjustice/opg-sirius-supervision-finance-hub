package allpay

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/stretchr/testify/assert"
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
	_, _ = client.newRequest(testContext(), "GET", "/", nil)
}

func testContext() context.Context {
	return auth.Context{
		Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("test")),
	}
}
