package api

import (
	"github.com/opg-sirius-finance-hub/finance-api/internal/validation"
	"github.com/opg-sirius-finance-hub/shared"
	"go.uber.org/zap"
	"net/http"
)

type Service interface {
	GetAccountInformation(id int) (*shared.AccountInformation, error)
	GetInvoices(id int) (*shared.Invoices, error)
	GetPermittedAdjustments(invoiceId int) ([]shared.AdjustmentType, error)
	GetFeeReductions(id int) (*shared.FeeReductions, error)
	CreateLedgerEntry(clientId int, invoiceId int, ledgerEntry *shared.CreateLedgerEntryRequest) (*shared.InvoiceReference, error)
	GetInvoiceAdjustments(id int) (*shared.InvoiceAdjustments, error)
	AddFeeReduction(id int, data shared.AddFeeReduction) error
	CancelFeeReduction(id int) error
	UpdatePendingInvoiceAdjustment(id int, status string) error
	AddManualInvoice(id int, invoice shared.AddManualInvoice) error
	GetBillingHistory(id int) ([]shared.BillingHistory, error)
}

type Server struct {
	Logger    *zap.SugaredLogger
	Service   Service
	Validator *validation.Validate
}

func (s *Server) SetupRoutes() {
	http.HandleFunc("GET /clients/{clientId}", s.getAccountInformation)
	http.HandleFunc("GET /clients/{clientId}/invoices", s.getInvoices)
	http.HandleFunc("GET /clients/{clientId}/invoices/{invoiceId}/permitted-adjustments", s.getPermittedAdjustments)
	http.HandleFunc("GET /clients/{clientId}/fee-reductions", s.getFeeReductions)
	http.HandleFunc("GET /clients/{clientId}/invoice-adjustments", s.getInvoiceAdjustments)
	http.HandleFunc("GET /clients/{clientId}/billing-history", s.getBillingHistory)

	http.HandleFunc("POST /clients/{clientId}/invoices", s.addManualInvoice)
	http.HandleFunc("POST /clients/{clientId}/invoices/{invoiceId}/ledger-entries", s.PostLedgerEntry)
	http.HandleFunc("PUT /clients/{clientId}/invoice-adjustments/{ledgerId}", s.updatePendingInvoiceAdjustment)
	http.HandleFunc("POST /clients/{clientId}/fee-reductions", s.addFeeReduction)
	http.HandleFunc("PUT /clients/{clientId}/fee-reductions/{feeReductionId}/cancel", s.cancelFeeReduction)

	http.Handle("/health-check", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
}
