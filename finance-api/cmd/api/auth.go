package api

import (
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) authenticate(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := auth.Context{Context: r.Context()}
		logger := telemetry.LoggerFromContext(ctx)

		authHeader := r.Header.Get("Authorization")
		requestToken := strings.Split(authHeader, "Bearer ")[1]
		token, err := s.JWT.Verify(requestToken)

		if err != nil {
			logger.Error("Unable to authorise user token: ", "err", err.Error())
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(*auth.Claims)
		userID, _ := strconv.Atoi(claims.ID)

		ctx.User = &shared.User{
			ID:    userID,
			Roles: claims.Roles,
		}

		h.ServeHTTP(w, r.WithContext(ctx))
	}
}
