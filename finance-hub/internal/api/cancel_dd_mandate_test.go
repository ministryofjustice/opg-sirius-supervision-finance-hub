package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"

	"github.com/stretchr/testify/assert"
)

func TestCancelDirectDebitMandate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/supervision-api/v1/clients/1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
			  "id": 1,
			  "firstname": "Account",
			  "surname": "Holder",
			  "caseRecNumber": "11111111",
			  "addressLine1": "1 Main Street",
			  "addressLine2": "Mainville",
			  "town": "Mainopolis",
			  "postcode": "MP1 2PM",
			  "feePayer": {
				"id": 1,
				"deputyStatus": "Active"
			  },
			  "activeCaseType": {
				"handle": "HW",
				"label": "Health & Welfare"
			  },
			  "clientStatus": {
				"handle": "ACTIVE",
				"label": "Active"
			  }
			}`))
		case "/clients/1/direct-debit":
			body, _ := io.ReadAll(r.Body)
			var data shared.CancelMandate
			if err := json.Unmarshal(body, &data); err != nil {
				t.Errorf("Invalid JSON body: %v", err)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{ts.URL + "", ts.URL + ""})

	err := client.CancelDirectDebitMandate(testContext(), 1)
	assert.Equal(t, nil, err)
}

func TestCancelDirectDebitMandate_GetPersonDetailsFails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{ts.URL + "", ts.URL + ""})

	err := client.CancelDirectDebitMandate(testContext(), 1)
	assert.Error(t, err)
}

func TestCancelDirectDebitMandate_CancelMandateFails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/supervision-api/v1/clients/1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
			  "id": 1,
			  "firstname": "Account",
			  "surname": "Holder",
			  "caseRecNumber": "11111111",
			  "addressLine1": "1 Main Street",
			  "addressLine2": "Mainville",
			  "town": "Mainopolis",
			  "postcode": "MP1 2PM",
			  "feePayer": {
				"id": 1,
				"deputyStatus": "Active"
			  },
			  "activeCaseType": {
				"handle": "HW",
				"label": "Health & Welfare"
			  },
			  "clientStatus": {
				"handle": "ACTIVE",
				"label": "Active"
			  }
			}`))
		case "/clients/1/direct-debit":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{ts.URL + "", ts.URL + ""})

	err := client.CancelDirectDebitMandate(testContext(), 1)
	assert.Error(t, err)
}
