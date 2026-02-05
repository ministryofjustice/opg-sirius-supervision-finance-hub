package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/stretchr/testify/assert"
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

func TestGetSession_contract(t *testing.T) {
	pact, err := consumer.NewV2Pact(consumer.MockHTTPProviderConfig{
		Consumer: "sirius-supervision-finance-hub",
		Provider: "sirius",
		LogDir:   "../../../logs",
		PactDir:  "../../../pacts",
	})
	assert.NoError(t, err)

	err = pact.
		AddInteraction().
		Given("User exists").
		UponReceiving("A request for the current user").
		WithRequest("GET", "/supervision-api/v1/users/current", func(b *consumer.V2RequestBuilder) {
			b.Header("Accept", matchers.S("application/json"))
		}).
		WillRespondWith(200, func(b *consumer.V2ResponseBuilder) {
			b.Header("Content-Type", matchers.S("application/json"))
			b.JSONBody(matchers.MapMatcher{
				"id":          matchers.Like(1),
				"displayName": matchers.Like("Colin Case"),
				"roles":       matchers.EachLike("Case Manager", 1),
			})
		}).
		ExecuteTest(t, func(config consumer.MockServerConfig) error {
			client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{fmt.Sprintf("http://%s:%d", config.Host, config.Port), ""})

			user, err := client.GetUserSession(testContext())
			if err != nil {
				return err
			}

			assert.EqualValues(t, &shared.User{
				ID:          1,
				DisplayName: "Colin Case",
				Roles:       []string{"Case Manager"},
			}, user)
			return nil
		})

	assert.NoError(t, err)
}
