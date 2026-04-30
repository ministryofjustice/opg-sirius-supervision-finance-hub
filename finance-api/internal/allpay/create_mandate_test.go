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
		if r.URL.Path != "/AllpayApi/Customers/SCHEME123/Mandates/Create" {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var req CreateMandateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("Invalid JSON body: %v", err)
		}
		if req.Customer.SchemeCode != schemeCode {
			t.Errorf("Expected schemeCode to be added, got %s", req.Customer.SchemeCode)
		}
		if req.Customer.Surname != "Doe" {
			t.Errorf("Expected surmane to have whitespace removed, got %s", req.Customer.Surname)
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
		Customer: Customer{
			ClientReference: "REF123",
			Surname:         " Doe ", // whitespace should be stripped before encoding
			Address: Address{
				Line1:    "1 Street",
				Town:     "Townville",
				PostCode: "TO1 2ST",
			},
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

func TestCreateMandate_DataFormatting(t *testing.T) {
	tests := []struct {
		name             string
		input            CreateMandateRequest
		expectedSurname  string
		expectedLine1    string
		expectedTown     string
		expectedPostCode string
	}{
		{
			name: "trims whitespace from all fields",
			input: CreateMandateRequest{Customer: Customer{
				Surname: "  Smith  ",
				Address: Address{Line1: "  1 High Street  ", Town: "  London  ", PostCode: "  SW1A 1AA  "},
			}},
			expectedSurname:  "Smith",
			expectedLine1:    "1 High Street",
			expectedTown:     "London",
			expectedPostCode: "SW1A 1AA",
		},
		{
			name: "truncates surname to 18 characters",
			input: CreateMandateRequest{Customer: Customer{
				Surname: "VeryLongSurnameExceedingLimit",
				Address: Address{},
			}},
			expectedSurname: "VeryLongSurnameExc",
		},
		{
			name: "truncates address line1 to 40 characters",
			input: CreateMandateRequest{Customer: Customer{
				Address: Address{Line1: "123 This Is A Very Long Street Name That Exceeds Forty Characters"},
			}},
			expectedLine1: "123 This Is A Very Long Street Name That",
		},
		{
			name: "truncates town to 40 characters",
			input: CreateMandateRequest{Customer: Customer{
				Address: Address{Town: "A Very Long Town Name That Clearly Exceeds The Forty Character Limit"},
			}},
			expectedTown: "A Very Long Town Name That Clearly Excee",
		},
		{
			name: "truncates postcode to 10 characters",
			input: CreateMandateRequest{Customer: Customer{
				Address: Address{PostCode: "SW1A 1AA TOOLONG"},
			}},
			expectedPostCode: "SW1A 1AA T",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured CreateMandateRequest

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &captured)
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			c := &Client{
				http: ts.Client(),
				Envs: Envs{apiHost: ts.URL, schemeCode: "SCHEME"},
			}

			_ = c.CreateMandate(testContext(), &tt.input)

			if tt.expectedSurname != "" && captured.Customer.Surname != tt.expectedSurname {
				t.Errorf("Surname: expected %q, got %q", tt.expectedSurname, captured.Customer.Surname)
			}
			if tt.expectedLine1 != "" && captured.Customer.Address.Line1 != tt.expectedLine1 {
				t.Errorf("Line1: expected %q, got %q", tt.expectedLine1, captured.Customer.Address.Line1)
			}
			if tt.expectedTown != "" && captured.Customer.Address.Town != tt.expectedTown {
				t.Errorf("Town: expected %q, got %q", tt.expectedTown, captured.Customer.Address.Town)
			}
			if tt.expectedPostCode != "" && captured.Customer.Address.PostCode != tt.expectedPostCode {
				t.Errorf("PostCode: expected %q, got %q", tt.expectedPostCode, captured.Customer.Address.PostCode)
			}
		})
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

	err := c.CreateMandate(testContext(), &CreateMandateRequest{})
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

	err := c.CreateMandate(testContext(), &CreateMandateRequest{})
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

	err := c.CreateMandate(testContext(), &CreateMandateRequest{})
	var validationErr ErrorValidation
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ErrorValidation, got %v", err)
	}
}
