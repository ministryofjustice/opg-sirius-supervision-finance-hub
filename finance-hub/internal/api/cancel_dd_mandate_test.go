package api

import (
	"encoding/json"
	"errors"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
		case "/clients/1/payment-method":
			body, _ := io.ReadAll(r.Body)
			var data shared.UpdatePaymentMethod
			if err := json.Unmarshal(body, &data); err != nil {
				t.Errorf("Invalid JSON body: %v", err)
			}
			if data.PaymentMethod != shared.PaymentMethodDemanded {
				t.Errorf("Expected payment method to be DEMANDED, got %s", data.PaymentMethod.Key())
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	mockAllPay := mockAllPayClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{SiriusURL: ts.URL, BackendURL: ts.URL}, &mockAllPay)

	err := client.CancelDirectDebitMandate(testContext(), 1)
	assert.Equal(t, nil, err)

	assert.True(t, mockAllPay.cancelMandateCalled)
}

func TestCancelDirectDebitMandate_GetPersonDetailsFails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	mockAllPay := mockAllPayClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{SiriusURL: ts.URL, BackendURL: ts.URL}, &mockAllPay)

	err := client.CancelDirectDebitMandate(testContext(), 1)
	assert.Error(t, err)
}

func TestCancelDirectDebitMandate_CancelMandateFails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	mockAllPay := mockAllPayClient{createMandateError: errors.New("server error")}
	client := NewClient(ts.Client(), &mockJWT, Envs{SiriusURL: ts.URL, BackendURL: ts.URL}, &mockAllPay)

	err := client.CancelDirectDebitMandate(testContext(), 1)
	assert.Error(t, err)
}

func TestCancelDirectDebitMandate_updatePaymentMethodFails(t *testing.T) {
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
		case "/clients/1/payment-method":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	mockAllPay := mockAllPayClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{SiriusURL: ts.URL, BackendURL: ts.URL}, &mockAllPay)

	logHandler := TestLogHandler{}
	err := client.CancelDirectDebitMandate(testContextWithLogger(&logHandler), 1)
	assert.Error(t, err)
	logHandler.assertLog(t, "failed to update payment method in Sirius after successful mandate cancellation in AllPay")
}
