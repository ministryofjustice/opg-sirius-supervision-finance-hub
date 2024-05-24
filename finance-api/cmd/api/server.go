package api

import (
	"github.com/ministryofjustice/opg-go-common/securityheaders"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/opg-sirius-finance-hub/finance-api/internal/validation"
	"github.com/opg-sirius-finance-hub/shared"
	"log/slog"
	"net/http"
)

type Service interface {
	GetAccountInformation(id int) (*shared.AccountInformation, error)
	GetInvoices(id int) (*shared.Invoices, error)
	GetFeeReductions(id int) (*shared.FeeReductions, error)
	CreateLedgerEntry(clientId int, invoiceId int, ledgerEntry *shared.CreateLedgerEntryRequest) error
	GetInvoiceAdjustments(id int) (*shared.InvoiceAdjustments, error)
	AddFeeReduction(id int, data shared.AddFeeReduction) error
	CancelFeeReduction(id int) error
	UpdatePendingInvoiceAdjustment(id int) error
}

type Server struct {
	Logger    *slog.Logger
	Service   Service
	Validator *validation.Validate
}

func (s *Server) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /clients/{clientId}", s.getAccountInformation)
	mux.HandleFunc("GET /clients/{clientId}/invoices", s.getInvoices)
	mux.HandleFunc("GET /clients/{clientId}/fee-reductions", s.getFeeReductions)
	mux.HandleFunc("GET /clients/{clientId}/invoice-adjustments", s.getInvoiceAdjustments)

	mux.HandleFunc("POST /clients/{clientId}/invoices/{invoiceId}/ledger-entries", s.PostLedgerEntry)
	mux.HandleFunc("PUT /clients/{clientId}/invoice-adjustments/{ledgerId}", s.updatePendingInvoiceAdjustment)
	mux.HandleFunc("POST /clients/{clientId}/fee-reductions", s.addFeeReduction)
	mux.HandleFunc("PUT /clients/{clientId}/fee-reductions/{feeReductionId}/cancel", s.cancelFeeReduction)
	mux.Handle("/health-check", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	return telemetry.Middleware(s.Logger)(securityheaders.Use(s.RequestLogger(mux)))
}

func (s *Server) RequestLogger(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Logger.Info(
			"application Request",
			"method", r.Method,
			"uri", r.URL.RequestURI(),
		)
		h.ServeHTTP(w, r)
	}
}
