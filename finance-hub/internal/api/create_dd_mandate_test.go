package api

import (
	"encoding/json"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateDirectDebitMandate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/supervision-api/v1/clients/1" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
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
		} else if r.URL.Path == "/clients/1/payment-method" {
			body, _ := io.ReadAll(r.Body)
			var data shared.UpdatePaymentMethod
			if err := json.Unmarshal(body, &data); err != nil {
				t.Errorf("Invalid JSON body: %v", err)
			}
			if data.PaymentMethod != shared.PaymentMethodDirectDebit {
				t.Errorf("Expected payment method to be DIRECT DEBIT, got %s", data.PaymentMethod.Key())
			}
			w.WriteHeader(http.StatusNoContent)
		} else {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	mockAllPay := mockAllPayClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{ts.URL + "", ts.URL + ""}, &mockAllPay)

	err := client.CreateDirectDebitMandate(testContext(), 1, AccountDetails{
		AccountName:   "Mrs Account Holder",
		AccountNumber: "12345678",
		SortCode:      "30-33-30",
	})
	assert.Equal(t, nil, err)
}

//TODO: Fail cases
//TODO: validator

//
//func TestAddDirectDebitx(t *testing.T) {
//	mockClient := SetUpTest()
//	mockJWT := mockJWTClient{}
//	mockAllPay := mockAllPayClient{}
//	client := NewClient(mockClient, &mockJWT, Envs{"http://localhost:3000", ""}, &mockAllPay)
//
//	jsonData := `{
//			"name": "Mrs Account Holder",
//			"sortCode": "30-33-30",
//			"account": "12345678",
//        }`
//
//	r := io.NopCloser(bytes.NewReader([]byte(jsonData)))
//
//	GetDoFunc = func(*http.Request) (*http.Response, error) {
//		return &http.Response{
//			StatusCode: 201,
//			Body:       r,
//		}, nil
//	}
//
//	err := client.CreateDirectDebitMandate(testContext(), 1, AccountDetails{
//		AccountName:   "Mrs Account Holder",
//		AccountNumber: "30-33-30",
//		SortCode:      "12345678",
//	})
//	assert.Equal(t, nil, err)
//}
//
//func TestAddDirectDebitUnauthorised(t *testing.T) {
//	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusUnauthorized)
//	}))
//	defer svr.Close()
//
//	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL}, nil)
//
//	err := client.CreateDirectDebitMandate(testContext(), 1, AccountDetails{
//		AccountName:   "Mrs Account Holder",
//		AccountNumber: "30-33-30",
//		SortCode:      "12345678",
//	})
//
//	assert.Equal(t, ErrUnauthorized.Error(), err.Error())
//}
//
//func TestAddDirectDebitReturns500Error(t *testing.T) {
//	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusInternalServerError)
//	}))
//	defer svr.Close()
//
//	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL}, nil)
//
//	err := client.CreateDirectDebitMandate(testContext(), 1, AccountDetails{
//		AccountName:   "Mrs Account Holder",
//		AccountNumber: "30-33-30",
//		SortCode:      "12345678",
//	})
//	assert.Equal(t, StatusError{
//		Code:   http.StatusInternalServerError,
//		URL:    svr.URL + "/clients/1/direct-debits",
//		Method: http.MethodPost,
//	}, err)
//}
//
//func TestAddDirectDebitReturnsValidationError(t *testing.T) {
//	validationErrors := apierror.ValidationError{
//		Errors: map[string]map[string]string{
//			"accountName": {
//				"required": "This field accountName needs to be looked at",
//			},
//		},
//	}
//	responseBody, _ := json.Marshal(validationErrors)
//	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusUnprocessableEntity)
//		_, _ = w.Write(responseBody)
//	}))
//	defer svr.Close()
//
//	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL}, nil)
//
//	err := client.CreateDirectDebitMandate(testContext(), 1, AccountDetails{
//		AccountName:   "Mrs Account Holder",
//		AccountNumber: "30-33-30",
//		SortCode:      "12345678",
//	})
//	expectedError := apierror.ValidationError{Errors: apierror.ValidationErrors{"accountName": map[string]string{"required": "This field accountName needs to be looked at"}}}
//	assert.Equal(t, expectedError, err.(apierror.ValidationError))
//}
