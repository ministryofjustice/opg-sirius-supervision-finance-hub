package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"golang.org/x/exp/maps"

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
		case "/clients/1/direct-debit":
			w.WriteHeader(http.StatusCreated)
			body, _ := io.ReadAll(r.Body)
			var data shared.CreateMandate
			if err := json.Unmarshal(body, &data); err != nil {
				t.Errorf("Invalid JSON body: %v", err)
			}
			assert.EqualValues(t, shared.CreateMandate{
				AllPayCustomer: shared.AllPayCustomer{
					ClientReference: "11111111",
					Surname:         "Holder",
				},
				Address: shared.Address{
					Line1:    "1 Main Street",
					Town:     "Mainopolis",
					PostCode: "MP1 2PM",
				},
				BankAccount: struct {
					BankDetails shared.AllPayBankDetails `json:"bankDetails"`
				}{
					BankDetails: shared.AllPayBankDetails{
						AccountName:   "Mrs Account Holder",
						SortCode:      "30-33-30",
						AccountNumber: "12345678",
					},
				},
			}, data)
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{ts.URL + "", ts.URL + ""})

	err := client.CreateDirectDebitMandate(testContext(), 1, AccountDetails{
		AccountName:   "Mrs Account Holder",
		AccountNumber: "12345678",
		SortCode:      "30-33-30",
	})
	assert.Equal(t, nil, err)
}

func TestCreateDirectDebitMandate_GetPersonDetailsFails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{ts.URL + "", ts.URL + ""})

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
	client := NewClient(ts.Client(), &mockJWT, Envs{ts.URL + "", ts.URL + ""})

	err := client.CreateDirectDebitMandate(testContext(), 1, AccountDetails{
		AccountName:   "Mrs Account Holder",
		AccountNumber: "12345678",
		SortCode:      "30-33-30",
	})
	assert.Error(t, err)
	expected := apierror.ValidationError{Errors: apierror.ValidationErrors{
		"ActiveOrder": map[string]string{
			"required": "",
		},
		"ClientStatus": map[string]string{
			"inactive": "",
		},
	}}
	assert.Equal(t, expected, err)
}

func TestCreateDirectDebitMandate_ModulusCheckFails(t *testing.T) {
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
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{
			  "field": "ModulusCheck",
			  "reason": "Failed"
			}`))
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{ts.URL + "", ts.URL + ""})

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

func TestCreateDirectDebitMandate_AllpayFails(t *testing.T) {
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
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{
			  "field": "Allpay",
			  "reason": "Failed"
			}`))
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{ts.URL + "", ts.URL + ""})

	err := client.CreateDirectDebitMandate(testContext(), 1, AccountDetails{
		AccountName:   "Mrs Account Holder",
		AccountNumber: "12345678",
		SortCode:      "30-33-30",
	})
	assert.Error(t, err)
	expected := apierror.ValidationError{Errors: apierror.ValidationErrors{
		"Allpay": map[string]string{
			"invalid": "",
		},
	}}
	assert.Equal(t, expected, err)
}

func TestCreateDirectDebitMandate_createMandateFails(t *testing.T) {
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

	err := client.CreateDirectDebitMandate(testContext(), 1, AccountDetails{
		AccountName:   "Mrs Account Holder",
		AccountNumber: "12345678",
		SortCode:      "30-33-30",
	})
	assert.Error(t, err)
}

func TestCreateDirectDebitMandate_createMandateFails_validation(t *testing.T) {
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
			w.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = w.Write([]byte(`{
			  "validation_errors": {
				"lastName": {
					"required": ""
				}
			  }
			}`))
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{ts.URL + "", ts.URL + ""})

	err := client.CreateDirectDebitMandate(testContext(), 1, AccountDetails{
		AccountName:   "Mrs Account Holder",
		AccountNumber: "12345678",
		SortCode:      "30-33-30",
	})
	assert.Error(t, err)
	expected := apierror.ValidationError{Errors: apierror.ValidationErrors{
		"lastName": map[string]string{
			"required": "",
		},
	}}
	assert.Equal(t, expected, err)
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

	c := &Client{}

	t.Run("success", func(t *testing.T) {
		errs := c.validateActiveClient(validClient)

		assert.Len(t, errs, 0)
	})

	t.Run("missing FeePayer", func(t *testing.T) {
		client := validClient
		client.FeePayer = nil
		errs := c.validateActiveClient(client)

		assert.NotNil(t, errs)
		assert.Equal(t, 1, len(maps.Keys(errs)))
		assert.Equal(t, "inactive", maps.Keys((errs)["FeePayer"])[0])
	})

	t.Run("inactive FeePayer", func(t *testing.T) {
		client := validClient
		client.FeePayer = &shared.FeePayer{Status: "Inactive"}
		errs := c.validateActiveClient(client)

		assert.NotNil(t, errs)
		assert.Equal(t, 1, len(maps.Keys(errs)))
		assert.Equal(t, "inactive", maps.Keys((errs)["FeePayer"])[0])
	})

	t.Run("missing ActiveCaseType", func(t *testing.T) {
		client := validClient
		client.ActiveCaseType = nil
		errs := c.validateActiveClient(client)

		assert.NotNil(t, errs)
		assert.Equal(t, 1, len(maps.Keys(errs)))
		assert.Equal(t, "required", maps.Keys((errs)["ActiveOrder"])[0])
	})

	t.Run("inactive ClientStatus", func(t *testing.T) {
		client := validClient
		client.ClientStatus = &shared.RefData{Handle: "INACTIVE"}
		errs := c.validateActiveClient(client)

		assert.NotNil(t, errs)
		assert.Equal(t, 1, len(maps.Keys(errs)))
		assert.Equal(t, "inactive", maps.Keys((errs)["ClientStatus"])[0])
	})

	t.Run("return multiple errors", func(t *testing.T) {
		client := validClient
		client.ClientStatus = &shared.RefData{Handle: "INACTIVE"}
		client.ActiveCaseType = nil
		errs := c.validateActiveClient(client)

		assert.NotNil(t, errs)
		assert.Equal(t, 2, len(maps.Keys(errs)))
	})
}
