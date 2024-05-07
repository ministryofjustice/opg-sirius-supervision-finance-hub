package api

import (
	"github.com/opg-sirius-finance-hub/auth"
	"github.com/opg-sirius-finance-hub/shared"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type Service interface {
	GetAccountInformation(id int) (*shared.AccountInformation, error)
	GetInvoices(id int) (*shared.Invoices, error)
	GetFeeReductions(id int) (*shared.FeeReductions, error)
}

type Server struct {
	Logger    *zap.SugaredLogger
	Service   Service
	JwtConfig auth.JwtConfig
}

func (s *Server) SetupRoutes() {
	http.Handle("GET /clients/{id}", s.jwtAuth(s.getAccountInformation))
	http.Handle("GET /clients/{id}/invoices", s.jwtAuth(s.getInvoices))
	http.Handle("GET /clients/{id}/fee-reductions", s.jwtAuth(s.getFeeReductions))
	http.Handle("/health-check", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
}

func (s *Server) jwtAuth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.JwtConfig.Enabled {
			authHeader := r.Header.Get("Authorization")
			requestToken := strings.Split(authHeader, "Bearer ")[1]
			token, err := s.JwtConfig.Verify(requestToken)

			if err != nil {
				s.Logger.Errorw("Error in token verification: ", err.Error())
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			if claims, ok := token.Claims.(*auth.Claims); ok {
				if !contains(claims.Roles, "urn:opg:sirius:private-finance-manager") {
					s.Logger.Errorw("Invalid user role")
					http.Error(w, "Invalid user role", http.StatusUnauthorized)
					return
				}
			}
			s.Logger.Infow("JWT successfully auth'd")
		}
		next.ServeHTTP(w, r)
	})
}

func contains(arr []string, v string) bool {
	for _, s := range arr {
		if s == v {
			return true
		}
	}
	return false
}
