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
	"strconv"
)

type ApiClient interface {
	GetCurrentUserDetails(sirius.Context) (model.Assignee, error)
	GetPersonDetails(sirius.Context, int) (model.Person, error)
	GetFeeReductions(sirius.Context, int) (model.FeeReductions, error)
	GetInvoices(sirius.Context, int) (model.Invoices, error)
	UpdateInvoice(sirius.Context, int, int, string, string) error
}

type router interface {
	Client() ApiClient
	execute(http.ResponseWriter, *http.Request, any) error
}

type Template interface {
	Execute(wr io.Writer, data any) error
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

func New(logger *zap.SugaredLogger, client ApiClient, templates map[string]*template.Template, envVars EnvironmentVars) http.Handler {
	wrap := wrapHandler(client, logger, templates["error.gotmpl"], envVars)

	mux := http.NewServeMux()

	mux.Handle("GET /clients/{id}/invoices", wrap(&InvoicesHandler{&route{client: client, tmpl: templates["invoices.gotmpl"], partial: "invoices"}}))
	mux.Handle("GET /clients/{id}/fee-reductions", wrap(&FeeReductionsHandler{&route{client: client, tmpl: templates["fee-reductions.gotmpl"], partial: "fee-reductions"}}))
	mux.Handle("GET /clients/{id}/pending-invoice-adjustments", wrap(&PendingInvoiceAdjustmentsHandler{&route{client: client, tmpl: templates["pending-invoice-adjustments.gotmpl"], partial: "pending-invoice-adjustments"}}))
	mux.Handle("GET /clients/{id}/invoices/{invoiceId}/ledger-entries", wrap(&UpdateInvoiceHandler{&route{client: client, tmpl: templates["update-invoice.gotmpl"], partial: "update-invoice"}}))
	mux.Handle("POST /clients/{id}/invoices/{invoiceId}/ledger-entries", wrap(&SubmitInvoiceHandler{&route{client: client, tmpl: templates["update-invoice.gotmpl"], partial: "error-summary"}}))

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

	clientId, _ := strconv.Atoi(r.PathValue("id"))

	return sirius.Context{
		Context:   r.Context(),
		Cookies:   r.Cookies(),
		XSRFToken: token,
		ClientId:  clientId,
	}
}
