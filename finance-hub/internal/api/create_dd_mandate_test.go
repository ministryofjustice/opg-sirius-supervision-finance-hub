package api

import (
	"encoding/json"
	"errors"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"golang.org/x/exp/maps"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateDirectDebitMandate(t *testing.T) {
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
			if data.PaymentMethod != shared.PaymentMethodDirectDebit {
				t.Errorf("Expected payment method to be DIRECT DEBIT, got %s", data.PaymentMethod.Key())
			}
			w.WriteHeader(http.StatusNoContent)
		default:
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

	assert.True(t, mockAllPay.modulusCalled)
	assert.True(t, mockAllPay.createMandateCalled)
}

func TestCreateDirectDebitMandate_GetPersonDetailsFails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
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
	assert.Error(t, err)
}

func TestCreateDirectDebitMandate_ValidationFails(t *testing.T) {
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
			  }
			}`))
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
	assert.Error(t, err)
}

func TestCreateDirectDebitMandate_ModulusCheckFails(t *testing.T) {
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
	mockAllPay := mockAllPayClient{modulusError: allpay.ErrorModulusCheckFailed{}}
	client := NewClient(ts.Client(), &mockJWT, Envs{ts.URL + "", ts.URL + ""}, &mockAllPay)

	err := client.CreateDirectDebitMandate(testContext(), 1, AccountDetails{
		AccountName:   "Mrs Account Holder",
		AccountNumber: "12345678",
		SortCode:      "30-33-30",
	})
	assert.Error(t, err)
	expected := apierror.ValidationError{Errors: apierror.ValidationErrors{
		"AccountDetails": map[string]string{
			"invalid": "",
		},
	}}
	assert.Equal(t, expected, err)
}

func TestCreateDirectDebitMandate_createMandateFails(t *testing.T) {
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
	client := NewClient(ts.Client(), &mockJWT, Envs{ts.URL + "", ts.URL + ""}, &mockAllPay)

	err := client.CreateDirectDebitMandate(testContext(), 1, AccountDetails{
		AccountName:   "Mrs Account Holder",
		AccountNumber: "12345678",
		SortCode:      "30-33-30",
	})
	assert.Error(t, err)
}

func TestCreateDirectDebitMandate_createMandateFails_validation(t *testing.T) {
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
	mockAllPay := mockAllPayClient{createMandateError: allpay.ErrorValidation{Messages: []string{"TitleInitials must not start with a comma", "PostCode - Only alphanumeric characters and spaces allowed"}}}
	client := NewClient(ts.Client(), &mockJWT, Envs{ts.URL + "", ts.URL + ""}, &mockAllPay)

	logHandler := TestLogHandler{}
	err := client.CreateDirectDebitMandate(testContextWithLogger(&logHandler), 1, AccountDetails{
		AccountName:   "Mrs Account Holder",
		AccountNumber: "12345678",
		SortCode:      "30-33-30",
	})
	assert.Error(t, err)

	logHandler.assertLog(t, "validation errors returned from allpay")
}

func TestCreateDirectDebitMandate_updatePaymentMethodFails(t *testing.T) {
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
	client := NewClient(ts.Client(), &mockJWT, Envs{ts.URL + "", ts.URL + ""}, &mockAllPay)

	logHandler := TestLogHandler{}
	err := client.CreateDirectDebitMandate(testContextWithLogger(&logHandler), 1, AccountDetails{
		AccountName:   "Mrs Account Holder",
		AccountNumber: "12345678",
		SortCode:      "30-33-30",
	})
	assert.Error(t, err)
	logHandler.assertLog(t, "failed to update payment method in Sirius after successful mandate creation in AllPay")
}

func TestValidateMandate(t *testing.T) {
	validClient := shared.Person{
		FeePayer:       &shared.FeePayer{Status: "Active"},
		ActiveCaseType: &shared.RefData{Handle: "H&W"},
		ClientStatus:   &shared.RefData{Handle: "ACTIVE"},
		CourtRef:       "1234567T",
		Surname:        "Smith",
		AddressLine1:   "123 Main Street",
		Town:           "Townsville",
		PostCode:       "TS1 1AA",
	}

	validDetails := AccountDetails{
		AccountName:   "John Smith",
		AccountNumber: "12345678",
		SortCode:      "11-22-33",
	}

	c := &Client{}

	t.Run("success", func(t *testing.T) {
		mandate, errs := c.validateMandate(validClient, validDetails)

		assert.Nil(t, errs)
		validMandate := allpay.CreateMandateRequest{
			ClientReference: "1234567T",
			Surname:         "Smith",
			Address: allpay.Address{
				Line1:    "123 Main Street",
				Town:     "Townsville",
				PostCode: "TS1 1AA",
			},
			BankAccount: struct {
				BankDetails allpay.BankDetails `json:"BankDetails"`
			}{
				BankDetails: allpay.BankDetails{
					AccountName:   "John Smith",
					SortCode:      "11-22-33",
					AccountNumber: "12345678",
				},
			},
		}
		assert.EqualValues(t, &validMandate, mandate)
	})

	t.Run("missing FeePayer", func(t *testing.T) {
		client := validClient
		client.FeePayer = nil
		_, errs := c.validateMandate(client, validDetails)

		assert.NotNil(t, errs)
		assert.Equal(t, 1, len(maps.Keys(*errs)))
		assert.Equal(t, "inactive", maps.Keys((*errs)["FeePayer"])[0])
	})

	t.Run("inactive FeePayer", func(t *testing.T) {
		client := validClient
		client.FeePayer = &shared.FeePayer{Status: "Inactive"}
		_, errs := c.validateMandate(client, validDetails)

		assert.NotNil(t, errs)
		assert.Equal(t, 1, len(maps.Keys(*errs)))
		assert.Equal(t, "inactive", maps.Keys((*errs)["FeePayer"])[0])
	})

	t.Run("missing ActiveCaseType", func(t *testing.T) {
		client := validClient
		client.ActiveCaseType = nil
		_, errs := c.validateMandate(client, validDetails)

		assert.NotNil(t, errs)
		assert.Equal(t, 1, len(maps.Keys(*errs)))
		assert.Equal(t, "required", maps.Keys((*errs)["ActiveOrder"])[0])
	})

	t.Run("inactive ClientStatus", func(t *testing.T) {
		client := validClient
		client.ClientStatus = &shared.RefData{Handle: "INACTIVE"}
		_, errs := c.validateMandate(client, validDetails)

		assert.NotNil(t, errs)
		assert.Equal(t, 1, len(maps.Keys(*errs)))
		assert.Equal(t, "inactive", maps.Keys((*errs)["ClientStatus"])[0])
	})

	t.Run("empty AccountName", func(t *testing.T) {
		details := validDetails
		details.AccountName = ""
		_, errs := c.validateMandate(validClient, details)

		assert.NotNil(t, errs)
		assert.Equal(t, 1, len(maps.Keys(*errs)))
		assert.Equal(t, "required", maps.Keys((*errs)["AccountName"])[0])
	})

	t.Run("long AccountName", func(t *testing.T) {
		details := validDetails
		details.AccountName = "ThisNameIsWayTooLongToBeValid"
		_, errs := c.validateMandate(validClient, details)

		assert.NotNil(t, errs)
		assert.Equal(t, 1, len(maps.Keys(*errs)))
		assert.Equal(t, "gteEighteen", maps.Keys((*errs)["AccountName"])[0])
	})

	t.Run("invalid SortCode", func(t *testing.T) {
		details := validDetails
		details.SortCode = "112233"
		_, errs := c.validateMandate(validClient, details)

		assert.NotNil(t, errs)
		assert.Equal(t, 1, len(maps.Keys(*errs)))
		assert.Equal(t, "len", maps.Keys((*errs)["SortCode"])[0])
	})

	t.Run("invalid AccountNumber", func(t *testing.T) {
		details := validDetails
		details.AccountNumber = "1234"
		_, errs := c.validateMandate(validClient, details)

		assert.NotNil(t, errs)
		assert.Equal(t, 1, len(maps.Keys(*errs)))
		assert.Equal(t, "len", maps.Keys((*errs)["AccountNumber"])[0])
	})

	t.Run("overlength AddressLine1", func(t *testing.T) {
		client := validClient
		client.AddressLine1 = strings.Repeat("A", 45)
		_, errs := c.validateMandate(client, validDetails)

		assert.NotNil(t, errs)
		assert.Equal(t, 1, len(maps.Keys(*errs)))
		assert.Equal(t, "required", maps.Keys((*errs)["AddressLine1"])[0])
	})

	t.Run("overlength Town", func(t *testing.T) {
		client := validClient
		client.Town = strings.Repeat("T", 45)
		_, errs := c.validateMandate(client, validDetails)

		assert.NotNil(t, errs)
		assert.Equal(t, 1, len(maps.Keys(*errs)))
		assert.Equal(t, "required", maps.Keys((*errs)["Town"])[0])
	})

	t.Run("overlength PostCode", func(t *testing.T) {
		client := validClient
		client.PostCode = strings.Repeat("P", 15)
		_, errs := c.validateMandate(client, validDetails)

		assert.NotNil(t, errs)
		assert.Equal(t, 1, len(maps.Keys(*errs)))
		assert.Equal(t, "required", maps.Keys((*errs)["PostCode"])[0])
	})

	t.Run("validate form fields first", func(t *testing.T) {
		client := validClient
		details := validDetails
		details.AccountNumber = "1234"
		client.PostCode = strings.Repeat("P", 15) // should not be validated as form fields take precedence
		_, errs := c.validateMandate(client, details)

		assert.NotNil(t, errs)
		assert.Equal(t, 1, len(maps.Keys(*errs)))
	})

	t.Run("return multiple errors", func(t *testing.T) {
		details := validDetails
		details.AccountNumber = "1234"
		details.SortCode = "AA-BB-CC"
		_, errs := c.validateMandate(validClient, details)

		assert.NotNil(t, errs)
		assert.Equal(t, 2, len(maps.Keys(*errs)))
	})
}
