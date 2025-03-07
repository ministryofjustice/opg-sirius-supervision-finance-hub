package auth

import (
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"net/http"
)

func XsrfCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
			ctx := r.Context().(Context)
			if r.FormValue("xsrfToken") != ctx.XSRFToken {
				logger := telemetry.LoggerFromContext(ctx)
				logger.Error("XSRF token mismatch")
				http.Error(w, "XSRF token mismatch", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
