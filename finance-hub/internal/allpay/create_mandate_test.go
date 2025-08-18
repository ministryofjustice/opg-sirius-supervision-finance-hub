package allpay

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateMandate_Success(t *testing.T) {
	schemeCode := "SCHEME123"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/AllpayApi/Customers/SCHEME123/VariableMandates/Create" {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var req CreateMandateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("Invalid JSON body: %v", err)
		}
		if req.SchemeCode != schemeCode {
			t.Errorf("Expected schemeCode to be added, got %s", schemeCode)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := &Client{
		http: ts.Client(),
		Envs: Envs{
			schemeCode: schemeCode,
			apiHost:    ts.URL,
		},
	}

	err := c.CreateMandate(testContext(), &CreateMandateRequest{
		ClientReference: "REF123",
		Surname:         "Doe",
		Address: Address{
			Line1:    "1 Street",
			Town:     "Townville",
			PostCode: "TO1 2ST",
		},
		BankAccount: struct {
			BankDetails BankDetails `json:"BankDetails"`
		}{
			BankDetails: BankDetails{
				AccountName:   "Donny Doe",
				SortCode:      "11-22-33",
				AccountNumber: "12345678",
			},
		},
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCreateMandate_RequestCreationFails(t *testing.T) {
	c := &Client{
		http: http.DefaultClient,
		Envs: Envs{
			schemeCode: "SCHEME123",
			apiHost:    "http://[::1]:namedport/", // Invalid URL
		},
	}

	err := c.CreateMandate(testContext(), &CreateMandateRequest{
		SchemeCode: "SCHEME123",
	})
	if err == nil {
		t.Error("Expected error due to request creation failure")
	}
}

func TestCreateMandate_UnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	c := &Client{
		http: ts.Client(),
		Envs: Envs{
			schemeCode: "SCHEME123",
			apiHost:    ts.URL,
		},
	}

	err := c.CreateMandate(testContext(), &CreateMandateRequest{
		SchemeCode: "SCHEME123",
	})
	if err == nil {
		t.Error("Expected error due to unexpected status code")
	}
}

func TestCreateMandate_ValidationErrorValidJSON(t *testing.T) {
	ve := ErrorValidation{Messages: []string{"Invalid data"}}
	body, _ := json.Marshal(ve)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write(body)
	}))
	defer ts.Close()

	c := &Client{
		http: ts.Client(),
		Envs: Envs{
			schemeCode: "SCHEME123",
			apiHost:    ts.URL,
		},
	}

	err := c.CreateMandate(testContext(), &CreateMandateRequest{
		SchemeCode: "SCHEME123",
	})
	var validationErr ErrorValidation
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ErrorValidation, got %v", err)
	}
}
