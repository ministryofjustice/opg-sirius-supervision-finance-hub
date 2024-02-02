package server

import (
	"github.com/ministryofjustice/opg-go-common/securityheaders"
	"github.com/opg-sirius-finance-hub/internal/sirius"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"html/template"
	"io"
	"net/http"
	"net/url"
)

type ApiClient interface {
	FinanceVarsClient
}

type Template interface {
	Execute(wr io.Writer, data any) error
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

func New(logger *zap.SugaredLogger, client ApiClient, templates map[string]*template.Template, envVars EnvironmentVars) http.Handler {
	wrap := wrapHandler(client, logger, templates["error.gotmpl"], envVars)

	mux := http.NewServeMux()

	mux.Handle("/invoices", wrap(invoices(client, templates["invoices.gotmpl"])))
	mux.Handle("/other", wrap(other(client, templates["other.gotmpl"])))

	mux.Handle("/", wrap(main(client, templates["index.gotmpl"])))

	mux.Handle("/health-check", healthCheck())

	static := http.FileServer(http.Dir(envVars.WebDir + "/static"))
	mux.Handle("/assets/", static)
	mux.Handle("/javascript/", static)
	mux.Handle("/stylesheets/", static)

	return otelhttp.NewHandler(http.StripPrefix(envVars.Prefix, securityheaders.Use(mux)), "supervision-finance")
}

func getContext(r *http.Request) sirius.Context {
	token := ""

	if r.Method == http.MethodGet {
		if cookie, err := r.Cookie("XSRF-TOKEN"); err == nil {
			token, _ = url.QueryUnescape(cookie.Value)
		}
	} else {
		token = r.FormValue("xsrfToken")
	}

	return sirius.Context{
		Context:   r.Context(),
		Cookies:   r.Cookies(),
		XSRFToken: token,
	}
}
