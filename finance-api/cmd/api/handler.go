package api

import (
	"errors"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"log/slog"
	"net/http"
)

type handlerFunc func(w http.ResponseWriter, r *http.Request) error

func (f handlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := f(w, r); err != nil {
		ctx := r.Context()
		logger := telemetry.LoggerFromContext(ctx)
		logger.Error("an api error occurred", slog.String("err", err.Error()))
		http.Error(w, err.Error(), HTTPStatus(err))
	}
}

func HTTPStatus(err error) int {
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
