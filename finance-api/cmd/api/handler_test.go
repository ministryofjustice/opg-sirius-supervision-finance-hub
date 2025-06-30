package api

import (
	"errors"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHttpStatus(t *testing.T) {
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
			assert.Equalf(t, tt.want, httpStatus(tt.err), "httpStatus(%v)", tt.err)
		})
	}
}

func Test_handlerFunc_ServeHTTP(t *testing.T) {
	tests := []struct {
		name       string
		f          handlerFunc
		statusCode int
		body       string
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
			body:       "error",
		},
		{
			name: "Error with body",
			f: func(w http.ResponseWriter, r *http.Request) error {
				return apierror.BadRequestError("A", "Something", nil)
			},
			statusCode: 400,
			body:       "{\"field\":\"A\",\"reason\":\"Something\"}",
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
			defer unchecked(res.Body.Close)

			assert.Equal(t, tt.statusCode, res.StatusCode)
			assert.Equal(t, tt.body, strings.TrimSpace(w.Body.String()))
		})
	}
}
