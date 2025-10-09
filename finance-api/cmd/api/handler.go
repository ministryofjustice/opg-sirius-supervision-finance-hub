package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"log/slog"
	"net/http"
	"runtime/debug"
)

type handlerFunc func(w http.ResponseWriter, r *http.Request) error

func (f handlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rcv := recover(); rcv != nil {
			logger := telemetry.LoggerFromContext(r.Context())
			logger.Error("panic recovered", slog.Any("error", rcv), slog.String("stack", string(debug.Stack())))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()

	if err := f(w, r); err != nil {
		var valErr apierror.ValidationError
		if !errors.As(err, &valErr) && !errors.Is(err, context.Canceled) {
			logger := telemetry.LoggerFromContext(r.Context())
			logger.Error("an api error occurred", slog.String("err", err.Error()))
		}
		writeError(w, err)
	}
}

func httpStatus(err error) int {
	if err == nil {
		return 0
	}
	var statusErr interface {
		error
		HTTPStatus() int
	}
	if errors.As(err, &statusErr) {
		return statusErr.HTTPStatus()
	}
	return http.StatusInternalServerError
}

func writeError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	statusCode := httpStatus(err)

	var withBodyErr interface {
		error
		HasData() bool
	}
	if errors.As(err, &withBodyErr) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(err)
	} else {
		http.Error(w, err.Error(), httpStatus(err))
	}
}
