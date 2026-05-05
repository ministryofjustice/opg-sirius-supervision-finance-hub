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

func TestTrimChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		limit    int
		expected string
	}{
		{"within limit", "hello", 10, "hello"},
		{"exactly at limit", "hello", 5, "hello"},
		{"exceeds limit", "hello world", 5, "hello"},
		{"trims leading and trailing spaces", "  hello  ", 10, "hello"},
		{"trims then truncates", "  hello world  ", 5, "hello"},
		{"unicode characters", "héllo", 4, "héll"},
		{"empty string", "", 5, ""},
		{"only spaces", "     ", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, trimChars(tt.input, tt.limit))
		})
	}
}

func testContext() context.Context {
	return auth.Context{
		Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("test")),
	}
}
