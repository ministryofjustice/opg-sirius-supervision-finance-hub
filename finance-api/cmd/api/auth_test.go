package api

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

type MockJWTClient struct {
	mock.Mock
}

func (m *MockJWTClient) Verify(requestToken string) (*jwt.Token, error) {
	args := m.Called(requestToken)
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func TestAuthenticate(t *testing.T) {
	mockJWTClient := &MockJWTClient{}
	server := &Server{JWT: mockJWTClient}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer valid-token")
	ctx := telemetry.ContextWithLogger(r.Context(), telemetry.NewLogger("test"))
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	mockJWTClient.On("Verify", "valid-token").Return(
		&jwt.Token{Claims: &auth.Claims{
			Roles:            []string{shared.RoleFinanceUser},
			RegisteredClaims: jwt.RegisteredClaims{ID: "1"},
		}}, nil)

	server.authenticate(handler).ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthenticateInvalidToken(t *testing.T) {
	mockJWTClient := &MockJWTClient{}
	server := &Server{JWT: mockJWTClient}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer invalid-token")
	ctx := telemetry.ContextWithLogger(r.Context(), telemetry.NewLogger("test"))
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	mockJWTClient.On("Verify", "invalid-token").Return(&jwt.Token{}, errors.New("invalid token"))

	server.authenticate(handler).ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthorise(t *testing.T) {
	server := &Server{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	ctx := auth.Context{
		Context: telemetry.ContextWithLogger(r.Context(), telemetry.NewLogger("test")),
		User: &shared.User{
			Roles: []string{shared.RoleFinanceUser},
		},
	}

	r = r.WithContext(ctx)

	server.authorise(shared.RoleFinanceUser)(handler).ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthoriseForbidden(t *testing.T) {
	server := &Server{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	ctx := auth.Context{
		Context: telemetry.ContextWithLogger(r.Context(), telemetry.NewLogger("test")),
		User: &shared.User{
			Roles: []string{shared.RoleFinanceUser},
		},
	}

	r = r.WithContext(ctx)

	server.authorise(shared.RoleFinanceManager)(handler).ServeHTTP(w, r)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
