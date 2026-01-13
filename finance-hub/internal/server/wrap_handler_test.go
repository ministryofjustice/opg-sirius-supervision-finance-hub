package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/api"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func TestStatusError_Code(t *testing.T) {
	assert.Equal(t, 0, StatusError(0).Code())
	assert.Equal(t, 200, StatusError(200).Code())
}

func TestStatusError_Error(t *testing.T) {
	assert.Equal(t, "0 ", StatusError(0).Error())
	assert.Equal(t, "200 OK", StatusError(200).Error())
	assert.Equal(t, "999 ", StatusError(999).Error())
}

type mockHandler struct {
	app    AppVars
	w      http.ResponseWriter
	r      *http.Request
	Err    error
	Called int
}

func (m *mockHandler) render(app AppVars, w http.ResponseWriter, r *http.Request) error {
	m.app = app
	m.w = w
	m.r = r
	m.Called = m.Called + 1
	return m.Err
}

func Test_wrapHandler_successful_request(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := auth.Context{
		Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("opg-sirius-supervision-finance-hub")),
		User:    &shared.User{ID: 1},
	}
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "test-url/1", nil)

	errorTemplate := &mockTemplate{}
	envVars := Envs{}
	nextHandlerFunc := wrapHandler(errorTemplate, "", envVars)
	next := &mockHandler{}
	httpHandler := nextHandlerFunc(next)
	httpHandler.ServeHTTP(w, r)

	assert.Nil(t, next.Err)
	assert.Equal(t, w, next.w)
	assert.Equal(t, r, next.r)
	assert.Equal(t, 1, next.Called)
	assert.Equal(t, "test-url/1", next.app.Path)
	assert.Equal(t, 200, w.Result().StatusCode)
}

func Test_wrapHandler_status_error_handling(t *testing.T) {
	tests := []struct {
		error     error
		wantCode  int
		wantError string
	}{
		{error: StatusError(400), wantCode: 400, wantError: "400 Bad Request"},
		{error: StatusError(401), wantCode: 401, wantError: "401 Unauthorized"},
		{error: StatusError(403), wantCode: 403, wantError: "403 Forbidden"},
		{error: StatusError(404), wantCode: 404, wantError: "404 Not Found"},
		{error: StatusError(500), wantCode: 500, wantError: "500 Internal Server Error"},
		{error: api.StatusError{Code: 400}, wantCode: 400, wantError: "  returned 400"},
		{error: api.StatusError{Code: 401}, wantCode: 401, wantError: "  returned 401"},
		{error: api.StatusError{Code: 403}, wantCode: 403, wantError: "  returned 403"},
		{error: api.StatusError{Code: 404}, wantCode: 404, wantError: "  returned 404"},
		{error: api.StatusError{Code: 500}, wantCode: 500, wantError: "  returned 500"},
		{error: context.Canceled, wantCode: 499, wantError: "context canceled"},
	}
	for i, test := range tests {
		t.Run("Scenario "+strconv.Itoa(i), func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx := auth.Context{
				Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("opg-sirius-supervision-finance-hub")),
				User:    &shared.User{ID: 1},
			}
			r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "test-url/1", nil)

			errorTemplate := &mockTemplate{error: errors.New("some template error")}
			envVars := Envs{}
			nextHandlerFunc := wrapHandler(errorTemplate, "", envVars)
			next := &mockHandler{Err: test.error}
			httpHandler := nextHandlerFunc(next)
			httpHandler.ServeHTTP(w, r)

			assert.Equal(t, 1, next.Called)
			assert.Equal(t, w, next.w)
			assert.Equal(t, r, next.r)
			assert.True(t, errorTemplate.executed)
			assert.IsType(t, ErrorVars{}, errorTemplate.lastVars)
			assert.Equal(t, test.wantCode, errorTemplate.lastVars.(ErrorVars).Code)
			assert.Equal(t, test.wantError, errorTemplate.lastVars.(ErrorVars).Error)
			assert.Equal(t, test.wantCode, w.Result().StatusCode)
		})
	}
}
