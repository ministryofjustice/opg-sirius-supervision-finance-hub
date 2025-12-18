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

func TestGetUser_cacheHit(t *testing.T) {
	mockClient := SetUpTest()
	mockJWT := mockJWTClient{}
	client := NewClient(mockClient, &mockJWT, Envs{"http://localhost:3000", "http://localhost:8181"})

	user := shared.User{
		ID:          1,
		DisplayName: "Tina Test",
		Roles:       []string{"Test Role"},
	}

	client.caches.updateUsers([]shared.User{user})

	expectedResponse := user

	user, err := client.GetUser(testContext(), 1)
	assert.Equal(t, expectedResponse, user)
	assert.Equal(t, nil, err)
}

func TestGetUser_cacheRefresh(t *testing.T) {
	mockClient := SetUpTest()
	mockJWT := mockJWTClient{}
	client := NewClient(mockClient, &mockJWT, Envs{"http://localhost:3000", "http://localhost:8181"})

	json := `[
				{
				   "id":1,
				   "name":"tina",
				   "phoneNumber":"12345678",
				   "user":[{
					  "displayName":"Finance Test",
					  "id":2
				   }],
				   "displayName":"Tina Test",
				   "deleted":false,
				   "email":"tina@opgtest.com",
				   "roles":[
					  "Test Role"
				   ],
				   "locked":false,
				   "suspended":false
				},
				{
				   "id":2,
				   "name":"case",
				   "phoneNumber":"12345678",
				   "user":[{
					  "displayName":"Lay Team 1 - (Supervision)",
					  "id":13
				   }],
				   "displayName":"case manager",
				   "deleted":false,
				   "email":"case.manager@opgtest.com",
				   "roles":[
					  "Case Manager"
				   ],
				   "locked":false,
				   "suspended":false
				}
			]`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	expectedResponse := shared.User{
		ID:          1,
		DisplayName: "Tina Test",
		Roles:       []string{"Test Role"},
	}

	user, err := client.GetUser(testContext(), 1)
	assert.Equal(t, expectedResponse, user)
	assert.Equal(t, nil, err)
}

func TestGetUser_cacheMiss(t *testing.T) {
	mockClient := SetUpTest()
	mockJWT := mockJWTClient{}
	client := NewClient(mockClient, &mockJWT, Envs{"http://localhost:3000", "http://localhost:8181"})

	json := `[]`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	expectedResponse := shared.User{
		ID:          0,
		DisplayName: "Unknown User",
		Roles:       nil,
	}

	user, err := client.GetUser(testContext(), 1)
	assert.Equal(t, expectedResponse, user)
	assert.Equal(t, nil, err)

	// the second time, the placeholder is set, so test to ensure it is fetched correctly
	user, err = client.GetUser(testContext(), 1)
	assert.Equal(t, expectedResponse, user)
	assert.Equal(t, nil, err)
}

func TestGetUserReturnsUnauthorisedClientError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})
	_, err := client.GetUser(testContext(), 1)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestGetUserReturns500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})

	_, err := client.GetUser(testContext(), 1)
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/supervision-api/v1/users",
		Method: http.MethodGet,
	}, err)
}

func TestGetUsers_contract(t *testing.T) {
	pact, err := consumer.NewV2Pact(consumer.MockHTTPProviderConfig{
		Consumer: "supervision-finance-hub",
		Provider: "sirius",
	})
	assert.NoError(t, err)

	err = pact.
		AddInteraction().
		Given("Users exists").
		UponReceiving("A request for users").
		WithRequest("GET", "/supervision-api/v1/users", func(b *consumer.V2RequestBuilder) {
			b.Header("Accept", matchers.S("application/json"))
		}).
		WillRespondWith(200, func(b *consumer.V2ResponseBuilder) {
			b.Header("Content-Type", matchers.S("application/json"))
			b.JSONBody(matchers.EachLike(matchers.MapMatcher{
				"id":          matchers.Like(1),
				"displayName": matchers.Like("Colin Case"),
				"roles":       matchers.EachLike("Case Manager", 1),
			}, 1))
		}).
		ExecuteTest(t, func(config consumer.MockServerConfig) error {
			client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{fmt.Sprintf("http://%s:%d", config.Host, config.Port), ""})

			user, err := client.GetUser(testContext(), 1)
			if err != nil {
				return err
			}

			assert.EqualValues(t, shared.User{
				ID:          1,
				DisplayName: "Colin Case",
				Roles:       []string{"Case Manager"},
			}, user)

			return nil
		})

	assert.NoError(t, err)
}
