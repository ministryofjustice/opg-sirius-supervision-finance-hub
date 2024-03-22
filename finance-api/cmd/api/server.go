package api

import (
	"github.com/opg-sirius-finance-hub/shared"
	"go.uber.org/zap"
	"net/http"
)

type Service interface {
	GetAccountInformation(id int) (*shared.AccountInformation, error)
}

type Server struct {
	Logger  *zap.SugaredLogger
	Service Service
}

func (s *Server) SetupRoutes() {
	http.HandleFunc("GET /clients/{id}", s.getAccountInformation)
}
