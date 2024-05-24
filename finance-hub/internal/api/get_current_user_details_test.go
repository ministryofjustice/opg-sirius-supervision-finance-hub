package api

import (
	"bytes"
	"github.com/opg-sirius-finance-hub/shared"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCurrentUserDetails(t *testing.T) {
	mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "http://localhost:8181")

	json := `{
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
			}`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	expectedResponse := shared.Assignee{
		Id:          65,
		DisplayName: "case manager",
		Roles:       []string{"Case Manager"},
	}

	teams, err := client.GetCurrentUserDetails(getContext(nil))
	assert.Equal(t, expectedResponse, teams)
	assert.Equal(t, nil, err)
}

func TestGetCurrentUserDetailsReturnsUnauthorisedClientError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, "")
	_, err := client.GetCurrentUserDetails(getContext(nil))
	assert.Equal(t, ErrUnauthorized, err)
}

func TestMyDetailsReturns500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, "")

	_, err := client.GetCurrentUserDetails(getContext(nil))
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/supervision-api/v1/users/current",
		Method: http.MethodGet,
	}, err)
}
