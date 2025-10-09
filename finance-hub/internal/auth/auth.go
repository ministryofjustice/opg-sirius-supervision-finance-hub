package auth

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
	"net/url"
)

type Context struct {
	context.Context
	Cookies   []*http.Cookie
	XSRFToken string
	User      *shared.User
}

func (c Context) WithContext(ctx context.Context) Context {
	return Context{
		Context:   ctx,
		Cookies:   c.Cookies,
		XSRFToken: c.XSRFToken,
		User:      c.User,
	}
}

func newContext(r *http.Request) Context {
	token := ""

	if cookie, err := r.Cookie("XSRF-TOKEN"); err == nil {
		token, _ = url.QueryUnescape(cookie.Value)
	}
	return Context{
		Context:   r.Context(),
		Cookies:   r.Cookies(),
		XSRFToken: token,
	}
}

type Client interface {
	GetUserSession(ctx context.Context) (*shared.User, error)
}

type EnvVars struct {
	SiriusPublicURL string
	Prefix          string
}

type Auth struct {
	Client  Client
	EnvVars EnvVars
}

func (a *Auth) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := newContext(r)
		logger := telemetry.LoggerFromContext(ctx)

		user, err := a.Client.GetUserSession(ctx)
		if err != nil {
			logger.Error("Error validating session.", "error", err)
			http.Redirect(w, r, a.redirectPath(r.URL.RequestURI()), http.StatusFound)
			return
		}
		if user == nil {
			logger.Info("User session not valid. Redirecting.")
			http.Redirect(w, r, a.redirectPath(r.URL.RequestURI()), http.StatusFound)
			return
		}

		ctx.User = user
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Auth) redirectPath(to string) string {
	return fmt.Sprintf("%s/auth?redirect=%s", a.EnvVars.SiriusPublicURL, url.QueryEscape(fmt.Sprintf("%s%s", a.EnvVars.Prefix, to)))
}
