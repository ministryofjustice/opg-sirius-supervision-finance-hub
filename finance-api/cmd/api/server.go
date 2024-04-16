package api

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/opg-sirius-finance-hub/auth"
	"github.com/opg-sirius-finance-hub/shared"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strings"
	"time"
)

type Service interface {
	GetAccountInformation(id int) (*shared.AccountInformation, error)
	GetInvoices(id int) (*shared.Invoices, error)
	GetFeeReductions(id int) (*shared.FeeReductions, error)
}

type Server struct {
	Logger  *zap.SugaredLogger
	Service Service
}

func (s *Server) SetupRoutes() {
	http.Handle("GET /clients/{id}", s.jwtAuth(s.getAccountInformation))
	http.Handle("GET /clients/{id}/invoices", s.jwtAuth(s.getInvoices))
	http.Handle("GET /clients/{id}/fee-reductions", s.jwtAuth(s.getFeeReductions))
	http.Handle("/health-check", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
}

func (s *Server) jwtAuth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		requestToken := strings.Split(authHeader, "Bearer ")[1]
		token, err := auth.Verify(requestToken, getEnv("JWT_SECRET", "mysupersecrettestkeythatis128bits"))

		if err != nil {
			s.Logger.Errorw("Error in token verification :", err.Error())
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		} else {
			claims := token.Claims.(jwt.MapClaims)
			var t *jwt.NumericDate
			if t, err = claims.GetExpirationTime(); err != nil || t.After(time.Now()) {
				s.Logger.Errorw("Token expired :", err.Error())
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
		}
		s.Logger.Infow("JWT successfully auth'd")
		next.ServeHTTP(w, r)
	})
}

func getEnv(key, def string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return def
}
