package api

import (
	"errors"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/opg-sirius-finance-hub/apierror"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "BadRequest",
			err:  apierror.BadRequest{},
			want: 400,
		},
		{
			name: "BadRequest",
			err:  apierror.BadRequests{},
			want: 400,
		},
		{
			name: "NotFound",
			err:  apierror.NotFound{},
			want: 404,
		},
		{
			name: "ValidationError",
			err:  apierror.ValidationError{},
			want: 422,
		},
		{
			name: "UnknownError",
			err:  errors.New("unknown error"),
			want: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, HTTPStatus(tt.err), "HTTPStatus(%v)", tt.err)
		})
	}
}

func Test_handlerFunc_ServeHTTP(t *testing.T) {
	tests := []struct {
		name       string
		f          handlerFunc
		statusCode int
	}{
		{
			name:       "OK",
			f:          func(w http.ResponseWriter, r *http.Request) error { return nil },
			statusCode: 200,
		},
		{
			name:       "Error",
			f:          func(w http.ResponseWriter, r *http.Request) error { return errors.New("error") },
			statusCode: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := telemetry.ContextWithLogger(r.Context(), telemetry.NewLogger("test"))
			r = r.WithContext(ctx)
			w := httptest.NewRecorder()

			tt.f.ServeHTTP(w, r)
			res := w.Result()

			assert.Equal(t, tt.statusCode, res.StatusCode)
		})
	}
}
