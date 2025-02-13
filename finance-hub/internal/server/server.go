package server

import (
	"context"
	"github.com/ministryofjustice/opg-go-common/securityheaders"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"html/template"
	"io"
	"log/slog"
	"net/http"
)

type ApiClient interface {
	GetCurrentUserDetails(context.Context) (shared.Assignee, error)
	GetPersonDetails(context.Context, int) (shared.Person, error)
	GetFeeReductions(context.Context, int) (shared.FeeReductions, error)
	GetInvoices(context.Context, int) (shared.Invoices, error)
	GetPermittedAdjustments(context.Context, int, int) ([]shared.AdjustmentType, error)
	AdjustInvoice(context.Context, int, int, int, string, string, string) error
	GetAccountInformation(context.Context, int) (shared.AccountInformation, error)
	GetInvoiceAdjustments(context.Context, int) (shared.InvoiceAdjustments, error)
	AddFeeReduction(context.Context, int, string, string, string, string, string) error
	CancelFeeReduction(context.Context, int, int, string) error
	UpdatePendingInvoiceAdjustment(context.Context, int, int, string) error
	AddManualInvoice(context.Context, int, string, *string, *string, *string, *string, *string, *string) error
	GetBillingHistory(context.Context, int) ([]shared.BillingHistory, error)
	GetUser(context.Context, int) (shared.Assignee, error)
	SubmitPaymentMethod(context.Context, int, string) error
}

type router interface {
	Client() ApiClient
	execute(http.ResponseWriter, *http.Request, any) error
	getClientID(r *http.Request) int
}

type Template interface {
	Execute(wr io.Writer, data any) error
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

type Envs struct {
	Port            string
	WebDir          string
	SiriusURL       string
	SiriusPublicURL string
	Prefix          string
	BackendURL      string
	BillingTeamID   int
}

func New(logger *slog.Logger, client ApiClient, templates map[string]*template.Template, envs Envs) http.Handler {
	wrap := wrapHandler(templates["error.gotmpl"], "main", envs)

	mux := http.NewServeMux()

	mux.Handle("GET /clients/{clientId}/invoices", wrap(&InvoicesHandler{&route{client: client, tmpl: templates["invoices.gotmpl"], partial: "invoices"}}))
	mux.Handle("GET /clients/{clientId}/fee-reductions", wrap(&FeeReductionsHandler{&route{client: client, tmpl: templates["fee-reductions.gotmpl"], partial: "fee-reductions"}}))
	mux.Handle("GET /clients/{clientId}/pending-invoice-adjustments", wrap(&PendingInvoiceAdjustmentsHandler{&route{client: client, tmpl: templates["pending-invoice-adjustments.gotmpl"], partial: "pending-invoice-adjustments"}}))
	mux.Handle("GET /clients/{clientId}/invoices/{invoiceId}/adjustments", wrap(&AdjustInvoiceFormHandler{&route{client: client, tmpl: templates["adjust-invoice.gotmpl"], partial: "adjust-invoice"}}))
	mux.Handle("GET /clients/{clientId}/fee-reductions/add", wrap(&UpdateFeeReductionHandler{&route{client: client, tmpl: templates["add-fee-reduction.gotmpl"], partial: "add-fee-reduction"}}))
	mux.Handle("GET /clients/{clientId}/payment-method/add", wrap(&PaymentMethodHandler{&route{client: client, tmpl: templates["set-up-payment-method.gotmpl"], partial: "set-up-payment-method"}}))
	mux.Handle("GET /clients/{clientId}/invoices/add", wrap(&UpdateManualInvoiceHandler{&route{client: client, tmpl: templates["add-manual-invoice.gotmpl"], partial: "add-manual-invoice"}}))
	mux.Handle("GET /clients/{clientId}/billing-history", wrap(&BillingHistoryHandler{&route{client: client, tmpl: templates["billing-history.gotmpl"], partial: "billing-history"}}))
	mux.Handle("GET /clients/{clientId}/fee-reductions/{feeReductionId}/cancel", wrap(&CancelFeeReductionHandler{&route{client: client, tmpl: templates["cancel-fee-reduction.gotmpl"], partial: "cancel-fee-reduction"}}))

	mux.Handle("POST /clients/{clientId}/invoices", wrap(&SubmitManualInvoiceHandler{&route{client: client, tmpl: templates["add-manual-invoice.gotmpl"], partial: "error-summary"}}))
	mux.Handle("POST /clients/{clientId}/invoices/{invoiceId}/adjustments", wrap(&SubmitInvoiceAdjustmentHandler{&route{client: client, tmpl: templates["adjust-invoice.gotmpl"], partial: "error-summary"}}))
	mux.Handle("POST /clients/{clientId}/fee-reductions/add", wrap(&SubmitFeeReductionsHandler{&route{client: client, tmpl: templates["add-fee-reduction.gotmpl"], partial: "error-summary"}}))
	mux.Handle("POST /clients/{clientId}/fee-reductions/{feeReductionId}/cancel", wrap(&SubmitCancelFeeReductionsHandler{&route{client: client, tmpl: templates["cancel-fee-reduction.gotmpl"], partial: "error-summary"}}))
	mux.Handle("POST /clients/{clientId}/pending-invoice-adjustments/{adjustmentId}/{adjustmentType}/{status}", wrap(&SubmitUpdatePendingInvoiceAdjustmentHandler{&route{client: client, tmpl: templates["pending-invoice-adjustments.gotmpl"], partial: "pending-invoice-adjustments"}}))
	mux.Handle("POST /clients/{clientId}/payment-method/add", wrap(&SubmitPaymentMethodHandler{&route{client: client, tmpl: templates["set-up-payment-method.gotmpl"], partial: "error-summary"}}))

	mux.Handle("/health-check", healthCheck())

	static := http.FileServer(http.Dir(envs.WebDir + "/static"))
	mux.Handle("/assets/", static)
	mux.Handle("/javascript/", static)
	mux.Handle("/stylesheets/", static)

	return otelhttp.NewHandler(http.StripPrefix(envs.Prefix, telemetry.Middleware(logger)(securityheaders.Use(mux))), "supervision-finance-hub")
}
