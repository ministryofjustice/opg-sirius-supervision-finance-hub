package allpay

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateClientDetails_Success(t *testing.T) {
	schemeCode := "SCHEME123"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		// ClientRef = base64("REF123") = "UkVGMTIz", Surname = base64("Smith") = "U21pdGg="
		if r.URL.Path != "/AllpayApi/Customers/SCHEME123/UkVGMTIz/U21pdGg=" {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var req updateClientDetailsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("Invalid JSON body: %v", err)
		}
		if req.Address.LastName != "Jones" {
			t.Errorf("Expected new surname Jones, got %s", req.Address.LastName)
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

	err := c.UpdateClientDetails(testContext(), &UpdateClientDetailsInput{
		ClientDetails: ClientDetails{
			ClientReference: "REF123",
			Surname:         " Smith ", // whitespace should be stripped before encoding
		},
		NewSurname: "Jones",
		Address: Address{
			Line1:    "1 Street",
			Town:     "Townville",
			PostCode: "TO1 2ST",
		},
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestUpdateClientDetails_DataFormatting(t *testing.T) {
	tests := []struct {
		name             string
		input            UpdateClientDetailsInput
		expectedSurname  string
		expectedLine1    string
		expectedTown     string
		expectedPostCode string
	}{
		{
			name: "trims whitespace from all fields",
			input: UpdateClientDetailsInput{
				ClientDetails: ClientDetails{Surname: "Smith"},
				NewSurname:    "  Jones  ",
				Address:       Address{Line1: "  1 High Street  ", Town: "  London  ", PostCode: "  SW1A 1AA  "},
			},
			expectedSurname:  "Jones",
			expectedLine1:    "1 High Street",
			expectedTown:     "London",
			expectedPostCode: "SW1A 1AA",
		},
		{
			name: "truncates new surname to 19 characters",
			input: UpdateClientDetailsInput{
				ClientDetails: ClientDetails{Surname: "Smith"},
				NewSurname:    "VeryLongSurnameExceedingLimit",
			},
			expectedSurname: "VeryLongSurnameExce",
		},
		{
			name: "truncates address line1 to 40 characters",
			input: UpdateClientDetailsInput{
				ClientDetails: ClientDetails{Surname: "Smith"},
				Address:       Address{Line1: "123 This Is A Very Long Street Name That Exceeds Forty Characters"},
			},
			expectedLine1: "123 This Is A Very Long Street Name That",
		},
		{
			name: "truncates town to 40 characters",
			input: UpdateClientDetailsInput{
				ClientDetails: ClientDetails{Surname: "Smith"},
				Address:       Address{Town: "A Very Long Town Name That Clearly Exceeds The Forty Character Limit"},
			},
			expectedTown: "A Very Long Town Name That Clearly Excee",
		},
		{
			name: "truncates postcode to 10 characters",
			input: UpdateClientDetailsInput{
				ClientDetails: ClientDetails{Surname: "Smith"},
				Address:       Address{PostCode: "SW1A 1AA TOOLONG"},
			},
			expectedPostCode: "SW1A 1AA T",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured updateClientDetailsRequest

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

			_ = c.UpdateClientDetails(testContext(), &tt.input)

			if tt.expectedSurname != "" && captured.Address.LastName != tt.expectedSurname {
				t.Errorf("LastName: expected %q, got %q", tt.expectedSurname, captured.Address.LastName)
			}
			if tt.expectedLine1 != "" && captured.Address.Line1 != tt.expectedLine1 {
				t.Errorf("Line1: expected %q, got %q", tt.expectedLine1, captured.Address.Line1)
			}
			if tt.expectedTown != "" && captured.Address.Town != tt.expectedTown {
				t.Errorf("Town: expected %q, got %q", tt.expectedTown, captured.Address.Town)
			}
			if tt.expectedPostCode != "" && captured.Address.PostCode != tt.expectedPostCode {
				t.Errorf("PostCode: expected %q, got %q", tt.expectedPostCode, captured.Address.PostCode)
			}
		})
	}
}

func TestUpdateClientDetails_RequestCreationFails(t *testing.T) {
	c := &Client{
		http: http.DefaultClient,
		Envs: Envs{
			schemeCode: "SCHEME123",
			apiHost:    "http://[::1]:namedport/", // Invalid URL
		},
	}

	err := c.UpdateClientDetails(testContext(), &UpdateClientDetailsInput{
		ClientDetails: ClientDetails{
			ClientReference: "REF123",
			Surname:         "Smith",
		},
	})
	if err == nil {
		t.Error("Expected error due to request creation failure")
	}
}

func TestUpdateClientDetails_UnexpectedStatus(t *testing.T) {
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

	err := c.UpdateClientDetails(testContext(), &UpdateClientDetailsInput{
		ClientDetails: ClientDetails{
			ClientReference: "REF123",
			Surname:         "Smith",
		},
	})
	if err == nil {
		t.Error("Expected error due to unexpected status code")
	}
}

func TestUpdateClientDetails_ValidationErrorValidJSON(t *testing.T) {
	ve := ErrorValidation{Messages: []string{"TitleInitials must not start with a comma"}}
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

	err := c.UpdateClientDetails(testContext(), &UpdateClientDetailsInput{
		ClientDetails: ClientDetails{
			ClientReference: "REF123",
			Surname:         "Smith",
		},
	})
	var validationErr ErrorValidation
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ErrorValidation, got %v", err)
	}
}
