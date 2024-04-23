package api

import (
	"github.com/opg-sirius-finance-hub/shared"
	"go.uber.org/zap"
	"net/http"
)

type Service interface {
	GetAccountInformation(id int) (*shared.AccountInformation, error)
	GetInvoices(id int) (*shared.Invoices, error)
	GetFeeReductions(id int) (*shared.FeeReductions, error)
	AddFeeReduction(body shared.AddFeeReduction) error
}

type Server struct {
	Logger  *zap.SugaredLogger
	Service Service
}

func (s *Server) SetupRoutes() {
	http.HandleFunc("GET /clients/{id}", s.getAccountInformation)
	http.HandleFunc("GET /clients/{id}/invoices", s.getInvoices)
	http.HandleFunc("GET /clients/{id}/fee-reductions", s.getFeeReductions)
	http.HandleFunc("POST /clients/{id}/fee-reductions", s.addFeeReduction)
	http.Handle("/health-check", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
}
