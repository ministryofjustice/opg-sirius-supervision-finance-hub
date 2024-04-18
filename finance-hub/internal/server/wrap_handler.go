package server

import (
	"errors"
	"fmt"
	"github.com/opg-sirius-finance-hub/auth"
	"github.com/opg-sirius-finance-hub/finance-hub/internal/api"
	"github.com/opg-sirius-finance-hub/finance-hub/internal/config"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type ErrorVars struct {
	Code  int
	Error string
	config.EnvironmentVars
}

type RedirectError string

func (e RedirectError) Error() string {
	return "redirect to " + string(e)
}

func (e RedirectError) To() string {
	return string(e)
}

type StatusError int

func (e StatusError) Error() string {
	code := e.Code()

	return fmt.Sprintf("%d %s", code, http.StatusText(code))
}

func (e StatusError) Code() int {
	return int(e)
}

type Handler interface {
	render(app AppVars, w http.ResponseWriter, r *http.Request) error
}

func wrapHandler(logger *zap.SugaredLogger, tmplError Template, envVars config.EnvironmentVars) func(next Handler) http.Handler {
	jwtConfig := auth.JwtConfig{Enabled: envVars.JwtEnabled, Secret: envVars.JwtSecret, Expiry: envVars.JwtExpiry}
	return func(next Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			if !isJwtAuthenticated(logger, r, jwtConfig) {
				http.Redirect(w, r, envVars.SiriusURL+"/auth", http.StatusFound)
				return
			}
			vars, err := NewAppVars(r, envVars)
			if err == nil {
				err = next.render(vars, w, r)
			}

			logger.Infow(
				"Application Request",
				"method", r.Method,
				"uri", r.URL.RequestURI(),
				"duration", time.Since(start),
			)

			if err != nil {
				if errors.Is(err, api.ErrUnauthorized) {
					http.Redirect(w, r, envVars.SiriusURL+"/auth", http.StatusFound)
					return
				}

				var redirect RedirectError
				if errors.As(err, &redirect) {
					http.Redirect(w, r, envVars.Prefix+redirect.To(), http.StatusFound)
					return
				}

				logger.Errorw("Error handler", err)

				code := http.StatusInternalServerError
				var serverStatusError StatusError
				if errors.As(err, &serverStatusError) {
					code = serverStatusError.Code()
				}
				var siriusStatusError api.StatusError
				if errors.As(err, &siriusStatusError) {
					code = siriusStatusError.Code
				}

				w.WriteHeader(code)
				errVars := ErrorVars{
					Code:            code,
					Error:           err.Error(),
					EnvironmentVars: envVars,
				}
				err = tmplError.Execute(w, errVars)

				if err != nil {
					logger.Errorw("Failed to render error template", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}
		})
	}
}

func isJwtAuthenticated(logger *zap.SugaredLogger, r *http.Request, jwtConfig auth.JwtConfig) bool {
	if !jwtConfig.Enabled {
		return true
	}

	cookie, _ := r.Cookie("OPG-TOKEN")
	if cookie == nil {
		logger.Errorw("Missing cookie token")
		return false
	}

	_, err := jwtConfig.Verify(cookie.Value)

	if err != nil {
		logger.Errorw("Error in token verification :", err.Error())
		return false
	}
	return true
}
