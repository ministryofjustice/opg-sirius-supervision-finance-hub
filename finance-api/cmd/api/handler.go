package api

import (
	"errors"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"net/http"
)

type handlerFunc func(w http.ResponseWriter, r *http.Request) error

func (f handlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := f(w, r); err != nil {
		ctx := r.Context()
		logger := telemetry.LoggerFromContext(ctx)
		logger.Error("something went wrong", err)
		if err != nil {
			http.Error(w, err.Error(), HTTPStatus(err))
		}
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
