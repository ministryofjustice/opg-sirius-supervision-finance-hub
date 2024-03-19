package api

import (
	"github.com/opg-sirius-finance-hub/finance-api/internal/service"
	"log"
	"net/http"
)

type Server struct {
	Logger  *log.Logger
	Service *service.Service
}

func (s *Server) SetupRoutes() {
	http.HandleFunc("GET /users/current", s.getCurrentUser)
}
