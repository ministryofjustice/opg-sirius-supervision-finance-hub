package api

import (
	"context"
	"github.com/ministryofjustice/opg-go-common/securityheaders"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/opg-sirius-finance-hub/finance-api/internal/validation"
	"github.com/opg-sirius-finance-hub/shared"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"log/slog"
	"net/http"
)

type Service interface {
	GetAccountInformation(ctx context.Context, id int) (*shared.AccountInformation, error)
	GetInvoices(ctx context.Context, clientId int) (*shared.Invoices, error)
	GetPermittedAdjustments(ctx context.Context, invoiceId int) ([]shared.AdjustmentType, error)
	GetFeeReductions(ctx context.Context, invoiceId int) (*shared.FeeReductions, error)
	AddInvoiceAdjustment(ctx context.Context, clientId int, invoiceId int, ledgerEntry *shared.AddInvoiceAdjustmentRequest) (*shared.InvoiceReference, error)
	GetInvoiceAdjustments(ctx context.Context, clientId int) (*shared.InvoiceAdjustments, error)
	AddFeeReduction(ctx context.Context, clientId int, data shared.AddFeeReduction) error
	CancelFeeReduction(ctx context.Context, id int, cancelledFeeReduction shared.CancelFeeReduction) error
	UpdatePendingInvoiceAdjustment(ctx context.Context, clientId int, adjustmentId int, status shared.AdjustmentStatus) error
	AddManualInvoice(ctx context.Context, clientId int, invoice shared.AddManualInvoice) error
	GetBillingHistory(ctx context.Context, id int) ([]shared.BillingHistory, error)
	ReapplyCredit(ctx context.Context, clientID int32) error
}

type Server struct {
	Service   Service
	Validator *validation.Validate
}

func (s *Server) SetupRoutes(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	// handleFunc is a replacement for mux.HandleFunc
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handleFunc := func(pattern string, h handlerFunc) {
		// Configure the "http.route" for the HTTP instrumentation.
		handler := otelhttp.WithRouteTag(pattern, h)
		mux.Handle(pattern, handler)
	}
	handleFunc("GET /clients/{clientId}", s.getAccountInformation)
	handleFunc("GET /clients/{clientId}/invoices", s.getInvoices)
	handleFunc("GET /clients/{clientId}/invoices/{invoiceId}/permitted-adjustments", s.getPermittedAdjustments)
	handleFunc("GET /clients/{clientId}/fee-reductions", s.getFeeReductions)
	handleFunc("GET /clients/{clientId}/invoice-adjustments", s.getInvoiceAdjustments)
	handleFunc("GET /clients/{clientId}/billing-history", s.getBillingHistory)

	handleFunc("POST /clients/{clientId}/invoices", s.addManualInvoice)
	handleFunc("POST /clients/{clientId}/invoices/{invoiceId}/invoice-adjustments", s.AddInvoiceAdjustment)
	handleFunc("PUT /clients/{clientId}/invoice-adjustments/{adjustmentId}", s.updatePendingInvoiceAdjustment)
	handleFunc("POST /clients/{clientId}/fee-reductions", s.addFeeReduction)
	handleFunc("PUT /clients/{clientId}/fee-reductions/{feeReductionId}/cancel", s.cancelFeeReduction)

	handleFunc("POST /events", s.handleEvents)

	handleFunc("/health-check", func(w http.ResponseWriter, r *http.Request) error { return nil })

	return otelhttp.NewHandler(telemetry.Middleware(logger)(securityheaders.Use(s.RequestLogger(mux))), "supervision-finance-api")
}

func (s *Server) RequestLogger(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		telemetry.LoggerFromContext(r.Context()).Info(
			"API Request",
			"method", r.Method,
			"uri", r.URL.RequestURI(),
		)
		h.ServeHTTP(w, r)
	}
}
