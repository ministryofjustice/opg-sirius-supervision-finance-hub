package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Server) authenticateAPI(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := auth.NewContext(r)
		logger := s.Logger(ctx)

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Error("Unable to authorise user token: ", "err", "missing bearer token")
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}

		requestToken := strings.Split(authHeader, "Bearer ")[1]
		token, err := s.JWT.Verify(requestToken)

		if err != nil {
			logger.Error("Unable to authorise user token: ", "err", err.Error())
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(*auth.Claims)
		userID, _ := strconv.ParseInt(claims.ID, 10, 32)

		ctx.User = &shared.User{
			ID:    int32(userID),
			Roles: claims.Roles,
		}

		h.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (s *Server) authenticateEvent(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := auth.NewContext(r)
		logger := s.Logger(ctx)

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			logger.Error("Unable to authorise event: ", "err", "missing bearer token")
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}

		apiKey := strings.Split(authHeader, "Bearer ")[1]
		if apiKey != s.envs.EventBridgeAPIKey {
			logger.Error("Unable to authorise event: ", "err", "invalid bearer token")
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return
		}

		ctx.User = &shared.User{
			ID: s.envs.SystemUserID,
		}

		h.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (s *Server) authorise(role string) func(http.Handler) http.HandlerFunc {
	return func(h http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context().(auth.Context)

			if !ctx.User.HasRole(role) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			h.ServeHTTP(w, r)
		}
	}
}
