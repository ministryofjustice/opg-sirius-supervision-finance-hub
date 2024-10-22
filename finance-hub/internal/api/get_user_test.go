package api

import (
	"bytes"
	"context"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/opg-sirius-finance-hub/shared"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUser_cacheHit(t *testing.T) {
	mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "http://localhost:8181")

	user := shared.Assignee{
		Id:          1,
		DisplayName: "Tina Test",
		Roles:       []string{"Test Role"},
	}

	client.caches.updateUsers([]shared.Assignee{user})

	expectedResponse := user

	ctx := telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("opg-sirius-finance-hub"))
	teams, err := client.GetUser(Context{Context: ctx}, 1)
	assert.Equal(t, expectedResponse, teams)
	assert.Equal(t, nil, err)
}

func TestGetUser_cacheRefresh(t *testing.T) {
	mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "http://localhost:8181")

	json := `[
				{
				   "id":1,
				   "name":"tina",
				   "phoneNumber":"12345678",
				   "teams":[{
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
				   "id":65,
				   "name":"case",
				   "phoneNumber":"12345678",
				   "teams":[{
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

	expectedResponse := shared.Assignee{
		Id:          1,
		DisplayName: "Tina Test",
		Roles:       []string{"Test Role"},
	}

	ctx := telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("opg-sirius-finance-hub"))
	teams, err := client.GetUser(Context{Context: ctx}, 1)
	assert.Equal(t, expectedResponse, teams)
	assert.Equal(t, nil, err)
}

func TestGetUser_unknownUser(t *testing.T) {
	mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "http://localhost:8181")

	json := `[
				{
				   "id":1,
				   "name":"tina",
				   "phoneNumber":"12345678",
				   "teams":[{
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
				}
			]`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	expectedResponse := shared.Assignee{
		Id:          99,
		DisplayName: "Unknown User",
		Roles:       nil,
	}

	ctx := telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("opg-sirius-finance-hub"))
	teams, err := client.GetUser(Context{Context: ctx}, 99)
	assert.Equal(t, expectedResponse, teams)
	assert.Equal(t, nil, err)
}

func TestGetUserReturnsUnauthorisedClientError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, "")
	ctx := telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("opg-sirius-finance-hub"))
	_, err := client.GetUser(Context{Context: ctx}, 1)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestGetUserReturns500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, "")

	ctx := telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("opg-sirius-finance-hub"))
	_, err := client.GetUser(Context{Context: ctx}, 1)
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/supervision-api/v1/users",
		Method: http.MethodGet,
	}, err)
}
