package api

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

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

type mockJWT struct {
	valid bool
}

func (m *mockJWT) Verify(token string) (*jwt.Token, error) {
	if !m.valid {
		return nil, jwt.ErrTokenMalformed
	}
	claims := auth.Claims{
		Roles: []string{"admin"},
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "1",
			Issuer:    "urn:opg:payments-hub",
			Audience:  jwt.ClaimStrings{"urn:opg:payments-api"},
			Subject:   "urn:opg:sirius:users:1",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 5)),
		},
	}
	return &jwt.Token{Claims: &claims}, nil
}

func TestAuthenticateAPI(t *testing.T) {
	var reqCtx auth.Context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		reqCtx = r.Context().(auth.Context)
	})

	tests := []struct {
		name           string
		authHeader     string
		valid          bool
		expectedStatus int
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer valid-token",
			valid:          true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid-token",
			valid:          false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Missing token",
			authHeader:     "",
			valid:          false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{
				JWT: &mockJWT{valid: tt.valid},
			}
			ctx := telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("test"))
			req, err := http.NewRequestWithContext(ctx, "GET", "/", nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", tt.authHeader)

			rr := httptest.NewRecorder()
			server.authenticateAPI(handler).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, 1, int(reqCtx.User.ID))
				assert.Equal(t, []string{"admin"}, reqCtx.User.Roles)
			}
		})
	}
}

func TestAuthenticateEvent(t *testing.T) {
	server := &Server{
		envs: &Envs{
			EventBridgeAPIKey: "valid-api-key",
			SystemUserID:      1,
		},
	}

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

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Valid API key",
			authHeader:     "Bearer valid-api-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid API key",
			authHeader:     "Bearer invalid-api-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Missing API key",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("test"))
			req, err := http.NewRequestWithContext(ctx, "GET", "/", nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", tt.authHeader)

			rr := httptest.NewRecorder()
			server.authenticateEvent(handler).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
