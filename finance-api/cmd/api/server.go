package api

import (
	"github.com/opg-sirius-finance-hub/finance-api/internal/service"
	"go.uber.org/zap"
	"net/http"
)

type Server struct {
	Logger  *zap.SugaredLogger
	Service *service.Service
}

func (s *Server) SetupRoutes() {
	http.HandleFunc("GET /clients/{id}", s.getAccountInformation)
}
