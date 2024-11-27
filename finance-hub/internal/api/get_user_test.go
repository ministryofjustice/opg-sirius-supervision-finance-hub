package api

import (
	"bytes"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
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

	user, err := client.GetUser(getContext(nil), 1)
	assert.Equal(t, expectedResponse, user)
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
				   "id":65,
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

	expectedResponse := shared.Assignee{
		Id:          1,
		DisplayName: "Tina Test",
		Roles:       []string{"Test Role"},
	}

	user, err := client.GetUser(getContext(nil), 1)
	assert.Equal(t, expectedResponse, user)
	assert.Equal(t, nil, err)
}

func TestGetUser_cacheMiss(t *testing.T) {
	mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "http://localhost:8181")

	json := `[]`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	expectedResponse := shared.Assignee{
		Id:          0,
		DisplayName: "Unknown User",
		Roles:       nil,
	}

	user, err := client.GetUser(getContext(nil), 1)
	assert.Equal(t, expectedResponse, user)
	assert.Equal(t, nil, err)

	// the second time, the placeholder is set, so test to ensure it is fetched correctly
	user, err = client.GetUser(getContext(nil), 1)
	assert.Equal(t, expectedResponse, user)
	assert.Equal(t, nil, err)
}

func TestGetUserReturnsUnauthorisedClientError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, "")
	_, err := client.GetUser(getContext(nil), 1)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestGetUserReturns500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, "")

	_, err := client.GetUser(getContext(nil), 1)
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/supervision-api/v1/users",
		Method: http.MethodGet,
	}, err)
}
