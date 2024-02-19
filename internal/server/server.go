package server

import (
	"github.com/ministryofjustice/opg-go-common/securityheaders"
	"github.com/opg-sirius-finance-hub/internal/model"
	"github.com/opg-sirius-finance-hub/internal/sirius"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"html/template"
	"io"
	"net/http"
	"net/url"
)

type ApiClient interface {
	GetCurrentUserDetails(sirius.Context) (model.Assignee, error)
	GetPersonDetails(sirius.Context, int) (model.Person, error)
}

type Template interface {
	Execute(wr io.Writer, data any) error
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

func New(logger *zap.SugaredLogger, client ApiClient, templates map[string]*template.Template, envVars EnvironmentVars) http.Handler {
	// middleware
	// - logging
	// - errors
	// - vars
	wrap := wrapHandler(client, logger, templates["error.gotmpl"], envVars)

	mux := http.NewServeMux()

	mux.Handle("/invoices", wrap(&InvoicesHandler{route{client: &client, tmpl: templates["invoices.gotmpl"], partial: "invoices"}}))
	mux.Handle("/other", wrap(&OtherHandler{route{client: &client, tmpl: templates["other.gotmpl"], partial: "other"}}))

	mux.HandleFunc("/", redirectToDefaultLanding)

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

// redirects to the invoices tab if root page is hit
// this isn't necessary if we have a different landing page
func redirectToDefaultLanding(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/invoices", http.StatusSeeOther)
}
