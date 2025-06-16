package server

import (
	"context"
	"github.com/ministryofjustice/opg-go-common/securityheaders"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/api"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"html/template"
	"io"
	"log/slog"
	"net/http"
)

type ApiClient interface {
	GetPersonDetails(context.Context, int) (shared.Person, error)
	GetFeeReductions(context.Context, int) (shared.FeeReductions, error)
	GetInvoices(context.Context, int) (shared.Invoices, error)
	GetPermittedAdjustments(context.Context, int, int) ([]shared.AdjustmentType, error)
	AddInvoiceAdjustment(context.Context, int, int, int, string, string, string, bool) error
	GetAccountInformation(context.Context, int) (shared.AccountInformation, error)
	GetInvoiceAdjustments(context.Context, int) (shared.InvoiceAdjustments, error)
	AddFeeReduction(context.Context, int, string, string, string, string, string) error
	CancelFeeReduction(context.Context, int, int, string) error
	UpdatePendingInvoiceAdjustment(context.Context, int, int, string) error
	AddManualInvoice(context.Context, int, string, *string, *string, *string, *string, *string, *string) error
	GetBillingHistory(context.Context, int) ([]shared.BillingHistory, error)
	SubmitPaymentMethod(context.Context, int, string) error
	GetUser(context.Context, int) (shared.User, error)
	GetRefunds(context.Context, int) (shared.Refunds, error)
	AddRefund(context.Context, int, string, string, string, string) error
	UpdateRefundDecision(context.Context, int, int, string) error
}

type router interface {
	Client() ApiClient
	execute(http.ResponseWriter, *http.Request, any) error
}

type Template interface {
	Execute(wr io.Writer, data any) error
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

type HtmxHandler interface {
	render(app AppVars, w http.ResponseWriter, r *http.Request) error
}

type Envs struct {
	Port            string
	WebDir          string
	SiriusURL       string
	SiriusPublicURL string
	Prefix          string
	BackendURL      string
	BillingTeamID   int
	ShowRefunds     bool
}

func New(logger *slog.Logger, client *api.Client, templates map[string]*template.Template, envs Envs) http.Handler {
	mux := http.NewServeMux()

	authenticator := auth.Auth{
		Client: client,
		EnvVars: auth.EnvVars{
			SiriusPublicURL: envs.SiriusPublicURL,
			Prefix:          envs.Prefix,
		},
	}

	handleMux := func(pattern string, h HtmxHandler) {
		errors := wrapHandler(templates["error.gotmpl"], "main", envs)
		mux.Handle(pattern, authenticator.Authenticate(auth.XsrfCheck(errors(h))))
	}

	handleMux("GET /clients/{clientId}/invoices", &InvoicesHandler{&route{client: client, tmpl: templates["invoices.gotmpl"], partial: "invoices"}})
	handleMux("GET /clients/{clientId}/fee-reductions", &FeeReductionsHandler{&route{client: client, tmpl: templates["fee-reductions.gotmpl"], partial: "fee-reductions"}})
	handleMux("GET /clients/{clientId}/invoice-adjustments", &InvoiceAdjustmentsHandler{&route{client: client, tmpl: templates["invoice-adjustments.gotmpl"], partial: "invoice-adjustments"}})
	handleMux("GET /clients/{clientId}/refunds", &RefundsHandler{&route{client: client, tmpl: templates["refunds.gotmpl"], partial: "refunds"}})
	handleMux("GET /clients/{clientId}/refunds/add", &AddRefundHandler{&route{client: client, tmpl: templates["add-refund.gotmpl"], partial: "add-refund"}})
	handleMux("GET /clients/{clientId}/invoices/{invoiceId}/adjustments", &AddInvoiceAdjustmentFormHandler{&route{client: client, tmpl: templates["adjust-invoice.gotmpl"], partial: "adjust-invoice"}})
	handleMux("GET /clients/{clientId}/fee-reductions/add", &UpdateFeeReductionHandler{&route{client: client, tmpl: templates["add-fee-reduction.gotmpl"], partial: "add-fee-reduction"}})
	handleMux("GET /clients/{clientId}/payment-method/add", &PaymentMethodHandler{&route{client: client, tmpl: templates["set-up-payment-method.gotmpl"], partial: "set-up-payment-method"}})
	handleMux("GET /clients/{clientId}/invoices/add", &AddManualInvoiceHandler{&route{client: client, tmpl: templates["add-manual-invoice.gotmpl"], partial: "add-manual-invoice"}})
	handleMux("GET /clients/{clientId}/billing-history", &BillingHistoryHandler{&route{client: client, tmpl: templates["billing-history.gotmpl"], partial: "billing-history"}})
	handleMux("GET /clients/{clientId}/fee-reductions/{feeReductionId}/cancel", &CancelFeeReductionHandler{&route{client: client, tmpl: templates["cancel-fee-reduction.gotmpl"], partial: "cancel-fee-reduction"}})

	handleMux("POST /clients/{clientId}/invoices", &SubmitManualInvoiceHandler{&route{client: client, tmpl: templates["add-manual-invoice.gotmpl"], partial: "error-summary"}})
	handleMux("POST /clients/{clientId}/invoices/{invoiceId}/adjustments", &SubmitInvoiceAdjustmentHandler{&route{client: client, tmpl: templates["adjust-invoice.gotmpl"], partial: "error-summary"}})
	handleMux("POST /clients/{clientId}/fee-reductions/add", &SubmitFeeReductionsHandler{&route{client: client, tmpl: templates["add-fee-reduction.gotmpl"], partial: "error-summary"}})
	handleMux("POST /clients/{clientId}/fee-reductions/{feeReductionId}/cancel", &SubmitCancelFeeReductionsHandler{&route{client: client, tmpl: templates["cancel-fee-reduction.gotmpl"], partial: "error-summary"}})
	handleMux("POST /clients/{clientId}/invoice-adjustments/{adjustmentId}/{adjustmentType}/{status}", &SubmitUpdatePendingInvoiceAdjustmentHandler{&route{client: client, tmpl: templates["invoice-adjustments.gotmpl"], partial: "invoice-adjustments"}})
	handleMux("POST /clients/{clientId}/payment-method/add", &SubmitPaymentMethodHandler{&route{client: client, tmpl: templates["set-up-direct-debit.gotmpl"], partial: "error-summary"}})
	handleMux("POST /clients/{clientId}/refunds", &SubmitRefundHandler{&route{client: client, tmpl: templates["add-refund.gotmpl"], partial: "error-summary"}})
	handleMux("POST /clients/{clientId}/refunds/{refundId}", &SubmitRefundDecisionHandler{&route{client: client, tmpl: templates["refunds.gotmpl"], partial: "refunds"}})

	mux.Handle("/health-check", healthCheck())

	static := http.FileServer(http.Dir(envs.WebDir + "/static"))
	mux.Handle("/assets/", static)
	mux.Handle("/javascript/", static)
	mux.Handle("/stylesheets/", static)

	return otelhttp.NewHandler(http.StripPrefix(envs.Prefix, telemetry.Middleware(logger)(securityheaders.Use(mux))), "supervision-finance-hub")
}
