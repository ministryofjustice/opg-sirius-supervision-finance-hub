package auth

import (
	"context"
	"errors"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockHandler struct {
	w      http.ResponseWriter
	r      *http.Request
	called bool
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.w = w
	m.r = r
	m.called = true
}

type mockAuthClient struct {
	User   *shared.User
	error  error
	called bool
}

func (m *mockAuthClient) GetUserSession(ctx context.Context) (*shared.User, error) {
	m.called = true
	return m.User, m.error
}

func Test_authenticate_success(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("test"))
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "test-url/1?q=abc", nil)
	r.AddCookie(&http.Cookie{Name: "XSRF-TOKEN", Value: "abcde"})
	r.AddCookie(&http.Cookie{Name: "session", Value: "12345"})

	user := &shared.User{
		ID: 1,
	}
	client := &mockAuthClient{User: user}

	auth := Auth{
		Client: client,
		EnvVars: EnvVars{
			SiriusPublicURL: "https://sirius.gov.uk",
		},
	}
	next := &mockHandler{}
	auth.Authenticate(next).ServeHTTP(w, r)

	assert.Equal(t, true, client.called)
	assert.Equal(t, w, next.w)
	assert.Equal(t, true, next.called)
	assert.Equal(t, 200, w.Result().StatusCode)

	rCtx := next.r.Context().(Context)
	assert.Equal(t, "12345", rCtx.Cookies[1].Value)
	assert.Equal(t, "abcde", rCtx.XSRFToken)
	assert.Equal(t, user, rCtx.User)
}

func Test_authenticate_unauthorised(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("test"))
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "test-url/1?q=abc", nil)

	client := &mockAuthClient{}

	auth := Auth{
		Client: client,
		EnvVars: EnvVars{
			SiriusPublicURL: "https://sirius.gov.uk",
			Prefix:          "finance-admin/",
		},
	}
	next := &mockHandler{}
	auth.Authenticate(next).ServeHTTP(w, r)

	assert.Equal(t, true, client.called)
	assert.Equal(t, false, next.called)
	assert.Equal(t, 302, w.Result().StatusCode)
	assert.Equal(t, "https://sirius.gov.uk/auth?redirect=finance-admin%2Ftest-url%2F1%3Fq%3Dabc", w.Result().Header.Get("Location"))
}

func Test_authenticate_error(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("test"))
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "test-url/1?q=abc", nil)

	client := &mockAuthClient{error: errors.New("something went wrong")}

	auth := Auth{
		Client: client,
		EnvVars: EnvVars{
			SiriusPublicURL: "https://sirius.gov.uk",
			Prefix:          "finance-admin/",
		},
	}
	next := &mockHandler{}
	auth.Authenticate(next).ServeHTTP(w, r)

	assert.Equal(t, true, client.called)
	assert.Equal(t, false, next.called)
	assert.Equal(t, 302, w.Result().StatusCode)
	assert.Equal(t, "https://sirius.gov.uk/auth?redirect=finance-admin%2Ftest-url%2F1%3Fq%3Dabc", w.Result().Header.Get("Location"))
}
