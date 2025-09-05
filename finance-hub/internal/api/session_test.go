package api

import (
	"bytes"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetUserSession(t *testing.T) {
	mockClient := &MockClient{}
	mockJwtClient := &mockJWTClient{}
	client := NewClient(mockClient, mockJwtClient, Envs{"http://localhost:3000", ""})

	json := `{
            "id": 1,
            "displayName": "Ian Test",
            "roles": ["role1","role2"]
        }`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	expected := shared.User{
		ID:          1,
		DisplayName: "Ian Test",
		Roles:       []string{"role1", "role2"},
	}

	user, err := client.GetUserSession(testContext())
	assert.Equal(t, &expected, user)
	assert.Equal(t, nil, err)
}

func TestClient_GetUserSession_Errors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   *shared.User
	}{
		{
			name:   "unauthorised",
			status: http.StatusUnauthorized,
			want:   nil,
		},
		{
			name:   "server error",
			status: http.StatusInternalServerError,
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer svr.Close()

			mockJwtClient := &mockJWTClient{}
			client := NewClient(http.DefaultClient, mockJwtClient, Envs{svr.URL, svr.URL})

			got, _ := client.GetUserSession(testContext())
			assert.Equalf(t, tt.want, got, "GetUserSession()")
		})
	}
}
