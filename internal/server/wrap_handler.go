package server

import (
	"errors"
	"fmt"
	"github.com/opg-sirius-finance-hub/internal/model"
	"github.com/opg-sirius-finance-hub/internal/sirius"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type ErrorVars struct {
	Code  int
	Error string
	EnvironmentVars
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

func isHxRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

type Handler interface {
	render(app PageVars, w http.ResponseWriter, r *http.Request) error
	replace(app AppVars, w http.ResponseWriter, r *http.Request) error
}

func wrapHandler(client ApiClient, logger *zap.SugaredLogger, tmplError Template, envVars EnvironmentVars) func(next Handler) http.Handler {
	return func(next Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			var err error
			vars := NewAppVars(r, envVars)
			if isHxRequest(r) {
				logger.Infow(
					"next - replace",
				)
				err = next.replace(*vars, w, r)
			} else {
				logger.Infow(
					"next - render",
				)
				// this data is only needed on a full page load, so fetch it here instead
				pageVars := PageVars{
					MyDetails: model.Assignee{},
					Client: ClientVars{
						FirstName:   "Ian",
						Surname:     "Moneybags",
						Outstanding: "1000000",
					},
					AppVars: *vars,
				}
				err = next.render(pageVars, w, r)
			}

			logger.Infow(
				"Application Request",
				"method", r.Method,
				"uri", r.URL.RequestURI(),
				"duration", time.Since(start),
			)

			if err != nil {
				if errors.Is(err, sirius.ErrUnauthorized) {
					http.Redirect(w, r, envVars.SiriusURL+"/auth", http.StatusFound)
					return
				}

				var redirect RedirectError
				if errors.As(err, &redirect) {
					http.Redirect(w, r, envVars.Prefix+"/"+redirect.To(), http.StatusFound)
					return
				}

				logger.Errorw("Error handler", err)

				code := http.StatusInternalServerError
				var serverStatusError StatusError
				if errors.As(err, &serverStatusError) {
					code = serverStatusError.Code()
				}
				var siriusStatusError sirius.StatusError
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
